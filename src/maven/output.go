package maven

import (
	"fmt"
	"io"
	"strings"
)

// PrintDependencies выводит все прямые зависимости, сохраняя порядок объявления.
// Отладочный список параметров из этапа 1 больше не выводится при обычном запуске.
func PrintDependencies(out io.Writer, name string, dependencies []Dependency) error {
	var text strings.Builder
	fmt.Fprintf(&text, "Прямые зависимости %s:\n", name)
	if len(dependencies) == 0 {
		text.WriteString("(нет прямых зависимостей)\n")
	}
	for _, dependency := range dependencies {
		info := "scope=" + dependency.Scope
		if dependency.Type != "jar" {
			info += ", type=" + dependency.Type
		}
		if dependency.Classifier != "" {
			info += ", classifier=" + dependency.Classifier
		}
		if dependency.Optional == "true" {
			info += ", optional=true"
		}
		fmt.Fprintf(&text, "%s [%s]\n", dependency.Coordinates, info)
	}
	_, err := io.WriteString(out, text.String())
	return err
}
