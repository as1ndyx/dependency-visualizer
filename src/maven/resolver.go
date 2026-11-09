package maven

import (
	"context"
	"fmt"
)

// resolver хранит POM только на время одного запроса прямых зависимостей.
type resolver struct {
	client       *Client
	repository   string
	cache        map[string]pom
	active       map[string]bool
	managedCache map[string]map[string]Dependency
}

// DirectDependencies получает прямые зависимости с учётом родителей, свойств и BOM.
// POM самих зависимостей не загружаются: транзитивный обход относится к этапу 3.
func (c *Client) DirectDependencies(ctx context.Context, repository, name string) ([]Dependency, error) {
	coordinates, err := ParseCoordinates(name)
	if err != nil {
		return nil, err
	}
	r := resolver{client: c, repository: repository, cache: make(map[string]pom),
		active: make(map[string]bool), managedCache: make(map[string]map[string]Dependency)}
	chain, err := r.parentChain(ctx, coordinates)
	if err != nil {
		return nil, err
	}
	managed, err := r.management(ctx, coordinates)
	if err != nil {
		return nil, err
	}
	return collectDependencies(chain, properties(chain, coordinates), managed)
}

// load повторно использует загруженный POM и ограничивает количество документов.
func (r *resolver) load(ctx context.Context, coordinates Coordinates) (pom, error) {
	key := coordinates.String()
	if model, ok := r.cache[key]; ok {
		return model, nil
	}
	if len(r.cache) >= MaxModels {
		return pom{}, fmt.Errorf("превышен предел %d POM родителей и BOM", MaxModels)
	}
	address, err := POMURL(r.repository, coordinates)
	if err != nil {
		return pom{}, err
	}
	model, err := r.client.fetch(ctx, address)
	if err != nil {
		return pom{}, err
	}
	r.cache[key] = model
	return model, nil
}

// parentChain возвращает POM от дочернего к корневому и обнаруживает циклы родителей.
func (r *resolver) parentChain(ctx context.Context, coordinates Coordinates) ([]pom, error) {
	var chain []pom
	seen := make(map[string]bool)
	for {
		if seen[coordinates.String()] {
			return nil, fmt.Errorf("цикл родительских POM: %s", coordinates)
		}
		seen[coordinates.String()] = true
		model, err := r.load(ctx, coordinates)
		if err != nil {
			return nil, err
		}
		chain = append(chain, model)
		if model.Parent == nil {
			return chain, nil
		}
		parent, err := expand(model.Parent.String(), properties([]pom{model}, coordinates), 0)
		if err != nil {
			return nil, err
		}
		coordinates, err = ParseCoordinates(parent)
		if err != nil {
			return nil, fmt.Errorf("parent: %w", err)
		}
	}
}

// collectDependencies объединяет объявления ребёнка и родителей, затем применяет версии.
func collectDependencies(chain []pom, values map[string]string, managed map[string]Dependency) ([]Dependency, error) {
	var result []Dependency
	positions := make(map[string]int)
	for _, model := range chain {
		for _, raw := range model.Dependencies {
			dependency, err := expandDependency(raw, values)
			if err != nil {
				return nil, err
			}
			key := dependency.key()
			if index, exists := positions[key]; exists {
				result[index] = result[index].inherit(dependency)
				continue
			}
			positions[key] = len(result)
			result = append(result, dependency)
		}
	}
	for i, dependency := range result {
		resolved, err := completeDependency(dependency.inherit(managed[dependency.key()]))
		if err != nil {
			return nil, err
		}
		result[i] = resolved
	}
	return result, nil
}

// completeDependency задаёт значения Maven по умолчанию и проверяет итоговые поля.
func completeDependency(dependency Dependency) (Dependency, error) {
	dependency = dependency.inherit(Dependency{Type: "jar", Scope: "compile", Optional: "false"})
	if _, err := ParseCoordinates(dependency.GroupID + ":" + dependency.ArtifactID + ":0"); err != nil {
		return Dependency{}, fmt.Errorf("неверное имя зависимости %s: %w", dependency.Coordinates, err)
	}
	if dependency.Version == "" {
		return Dependency{}, fmt.Errorf("не удалось определить версию зависимости %s", dependency.Coordinates)
	}
	switch dependency.Scope {
	case "compile", "provided", "runtime", "test", "system":
	default:
		return Dependency{}, fmt.Errorf("неизвестный scope %q у %s", dependency.Scope, dependency.Coordinates)
	}
	if dependency.Optional != "true" && dependency.Optional != "false" {
		return Dependency{}, fmt.Errorf("неверное optional %q у %s", dependency.Optional, dependency.Coordinates)
	}
	return dependency, nil
}
