package maven

import (
	"context"
	"fmt"
)

// management собирает таблицу версий и обнаруживает циклический импорт BOM.
func (r *resolver) management(ctx context.Context, coordinates Coordinates) (map[string]Dependency, error) {
	key := coordinates.String()
	if cached, ok := r.managedCache[key]; ok {
		return cached, nil
	}
	if r.active[key] {
		return nil, fmt.Errorf("цикл импорта BOM: %s", key)
	}
	r.active[key] = true
	defer delete(r.active, key)
	chain, err := r.parentChain(ctx, coordinates)
	if err != nil {
		return nil, err
	}
	managed, imports, err := managementEntries(chain, properties(chain, coordinates))
	if err != nil {
		return nil, err
	}
	for _, entry := range imports {
		imported, err := r.management(ctx, entry.Coordinates)
		if err != nil {
			return nil, fmt.Errorf("BOM %s: %w", entry.Coordinates, err)
		}
		for name, dependency := range imported {
			if _, exists := managed[name]; !exists {
				managed[name] = dependency
			}
		}
	}
	r.managedCache[key] = managed
	return managed, nil
}

// managementEntries отделяет обычные записи от BOM: сами BOM не являются зависимостями.
// Явные записи имеют приоритет над импортом; среди импортов побеждает первый.
func managementEntries(chain []pom, values map[string]string) (map[string]Dependency, []Dependency, error) {
	managed := make(map[string]Dependency)
	var imports []Dependency
	for _, model := range chain {
		for _, raw := range model.Managed {
			entry, err := expandDependency(raw, values)
			if err != nil {
				return nil, nil, err
			}
			if entry.Scope == "import" {
				if entry.Type != "pom" {
					return nil, nil, fmt.Errorf("импорт BOM должен иметь type=pom")
				}
				imports = append(imports, entry)
				continue
			}
			key := entry.key()
			if previous, exists := managed[key]; exists {
				managed[key] = previous.inherit(entry)
			} else {
				managed[key] = entry
			}
		}
	}
	return managed, imports, nil
}
