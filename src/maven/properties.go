package maven

import (
	"fmt"
	"regexp"
	"strings"
)

const maxPropertyDepth = 32

var propertyPattern = regexp.MustCompile(`\$\{([^{}]+)\}`)

// properties объединяет свойства родителей и дочернего POM.
// Значения наследуются до подстановки, поэтому ребёнок может переопределить версию.
func properties(chain []pom, coordinates Coordinates) map[string]string {
	values := make(map[string]string)
	for i := len(chain) - 1; i >= 0; i-- {
		for _, property := range chain[i].Properties {
			values[property.XMLName.Local] = strings.TrimSpace(property.Value)
		}
	}
	for _, prefix := range []string{"project.", "pom.", ""} {
		values[prefix+"groupId"] = coordinates.GroupID
		values[prefix+"artifactId"] = coordinates.ArtifactID
		values[prefix+"version"] = coordinates.Version
	}
	if parent := chain[0].Parent; parent != nil {
		for _, prefix := range []string{"project.parent.", "pom.parent.", "parent."} {
			values[prefix+"groupId"] = parent.GroupID
			values[prefix+"artifactId"] = parent.ArtifactID
			values[prefix+"version"] = parent.Version
		}
	}
	return values
}

// expand рекурсивно подставляет свойства и сообщает о неизвестных ссылках и циклах.
func expand(value string, values map[string]string, depth int) (string, error) {
	if depth > maxPropertyDepth {
		return "", fmt.Errorf("цикл или слишком большая глубина свойств POM: %s", value)
	}
	var resultErr error
	expandedBytes := 0
	result := propertyPattern.ReplaceAllStringFunc(value, func(match string) string {
		if resultErr != nil {
			return match
		}
		name := match[2 : len(match)-1]
		replacement, ok := values[name]
		if !ok {
			resultErr = fmt.Errorf("неизвестное свойство POM %s", match)
			return match
		}
		resolved, err := expand(replacement, values, depth+1)
		if err != nil {
			resultErr = err
		}
		expandedBytes += len(resolved)
		if expandedBytes > MaxPOMBytes {
			resultErr = fmt.Errorf("слишком большой результат подстановки свойства POM")
			return ""
		}
		return resolved
	})
	if len(result) > MaxPOMBytes {
		return "", fmt.Errorf("слишком большой результат подстановки свойства POM")
	}
	if resultErr == nil && strings.Contains(result, "${") {
		return "", fmt.Errorf("незавершённая ссылка на свойство POM")
	}
	return strings.TrimSpace(result), resultErr
}

// expandDependency подставляет свойства во всех полях зависимости.
func expandDependency(dependency Dependency, values map[string]string) (Dependency, error) {
	fields := []*string{&dependency.GroupID, &dependency.ArtifactID, &dependency.Version,
		&dependency.Scope, &dependency.Type, &dependency.Classifier, &dependency.Optional}
	for _, field := range fields {
		value, err := expand(*field, values, 0)
		if err != nil {
			return Dependency{}, err
		}
		*field = value
	}
	return dependency, nil
}
