// Package maven получает прямые зависимости из POM средствами стандартной библиотеки Go.
package maven

import (
	"fmt"
	"regexp"
	"strings"
)

// Coordinates — адрес пакета Maven: группа, имя артефакта и версия.
type Coordinates struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
}

var coordinatePattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]*:` +
	`[A-Za-z0-9_][A-Za-z0-9_.-]*:[A-Za-z0-9_][A-Za-z0-9_.+-]*$`)

// ParseCoordinates проверяет координаты перед построением URL POM.
func ParseCoordinates(value string) (Coordinates, error) {
	if !coordinatePattern.MatchString(value) {
		return Coordinates{}, fmt.Errorf("ожидаются координаты Maven groupId:artifactId:version")
	}
	parts := strings.Split(value, ":")
	for _, part := range strings.Split(parts[0], ".") {
		if part == "" {
			return Coordinates{}, fmt.Errorf("groupId содержит пустую часть")
		}
	}
	return Coordinates{parts[0], parts[1], parts[2]}, nil
}

// String возвращает координаты в формате groupId:artifactId:version.
func (c Coordinates) String() string {
	return c.GroupID + ":" + c.ArtifactID + ":" + c.Version
}

// Dependency — прямая зависимость и её параметры из POM.
// На этапе 2 показываются все области видимости, включая test и optional.
type Dependency struct {
	Coordinates
	Scope      string `xml:"scope"`
	Type       string `xml:"type"`
	Classifier string `xml:"classifier"`
	Optional   string `xml:"optional"`
}

// key различает зависимости одного артефакта с разными type и classifier.
func (d Dependency) key() string {
	kind := d.Type
	if kind == "" {
		kind = "jar"
	}
	return d.GroupID + ":" + d.ArtifactID + ":" + kind + ":" + d.Classifier
}

// inherit заполняет отсутствующие поля значениями родителя или dependencyManagement.
func (d Dependency) inherit(base Dependency) Dependency {
	fields := []struct {
		target *string
		value  string
	}{
		{&d.Version, base.Version}, {&d.Scope, base.Scope}, {&d.Type, base.Type},
		{&d.Classifier, base.Classifier}, {&d.Optional, base.Optional},
	}
	for _, field := range fields {
		if *field.target == "" {
			*field.target = field.value
		}
	}
	return d
}
