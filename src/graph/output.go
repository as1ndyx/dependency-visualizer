package graph

import (
	"fmt"
	"io"
	"strings"
)

// Print выводит список смежности в порядке DFS, обратные рёбра циклов и исключённые пакеты.
// На этапе 3 граф показывается текстом; язык Graphviz будет добавлен на этапе 5.
func Print(out io.Writer, graph Graph) error {
	var text strings.Builder
	fmt.Fprintf(&text, "Граф зависимостей %s (DFS):\n", graph.Root)
	if len(graph.Nodes) == 0 {
		text.WriteString("(корневой пакет исключён фильтром)\n")
	}
	for _, node := range graph.Nodes {
		dependencies := strings.Join(graph.Edges[node], ", ")
		if dependencies == "" {
			dependencies = "(нет зависимостей)"
		}
		fmt.Fprintf(&text, "%s -> %s\n", node, dependencies)
	}
	if len(graph.Cycles) == 0 {
		text.WriteString("Циклы: нет\n")
	} else {
		text.WriteString("Циклы (обратные рёбра DFS):\n")
		for _, edge := range graph.Cycles {
			fmt.Fprintf(&text, "%s -> %s\n", edge.From, edge.To)
		}
	}
	if len(graph.Excluded) != 0 {
		fmt.Fprintf(&text, "Исключены фильтром: %s\n", strings.Join(graph.Excluded, ", "))
	}
	_, err := io.WriteString(out, text.String())
	return err
}
