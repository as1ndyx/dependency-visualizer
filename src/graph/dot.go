package graph

import (
	"fmt"
	"io"
	"strings"
)

// PrintDOT выводит готовое описание Graphviz без служебного текста в stdout.
// Этап 5: корень выделяется синим цветом, обратные рёбра DFS — красным.
func PrintDOT(out io.Writer, graph Graph) error {
	var text strings.Builder
	title := "Зависимости " + graph.Root
	if len(graph.Nodes) == 0 {
		title += "\nКорневой пакет исключён фильтром"
	}
	if len(graph.Excluded) != 0 {
		title += "\nИсключены: " + strings.Join(graph.Excluded, ", ")
	}
	text.WriteString("digraph dependencies {\n")
	fmt.Fprintf(&text, "  graph [rankdir=LR, charset=\"UTF-8\", label=%s, labelloc=t];\n", quoteDOT(title))
	text.WriteString("  node [shape=box, style=\"rounded,filled\", fillcolor=\"#f8fafc\", fontname=\"Arial\"];\n")
	text.WriteString("  edge [color=\"#64748b\"];\n")
	writeDOTNodes(&text, graph)
	writeDOTEdges(&text, graph)
	text.WriteString("}\n")
	_, err := io.WriteString(out, text.String())
	return err
}

// writeDOTNodes объявляет все вершины, включая одиночный корень и листья.
func writeDOTNodes(text *strings.Builder, graph Graph) {
	for _, name := range graph.Nodes {
		label := strings.ReplaceAll(name, ":", "\n")
		fmt.Fprintf(text, "  %s [label=%s", quoteDOT(name), quoteDOT(label))
		if name == graph.Root {
			text.WriteString(", fillcolor=\"#dbeafe\", color=\"#2563eb\", penwidth=2")
		}
		text.WriteString("];\n")
	}
}

// writeDOTEdges сохраняет направления связей, общие зависимости и циклы.
func writeDOTEdges(text *strings.Builder, graph Graph) {
	cycles := make(map[Edge]bool)
	for _, edge := range graph.Cycles {
		cycles[edge] = true
	}
	for _, from := range graph.Nodes {
		for _, to := range graph.Edges[from] {
			fmt.Fprintf(text, "  %s -> %s", quoteDOT(from), quoteDOT(to))
			if cycles[Edge{From: from, To: to}] {
				text.WriteString(" [color=\"#dc2626\", penwidth=2]")
			}
			text.WriteString(";\n")
		}
	}
}

// quoteDOT экранирует кавычки, обратные косые черты и переносы в строках DOT.
func quoteDOT(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	return "\"" + replacer.Replace(value) + "\""
}
