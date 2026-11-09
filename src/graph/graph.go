// Package graph строит граф рекурсивным DFS и находит обратные зависимости на этапах 3–4.
package graph

import (
	"context"
	"fmt"
	"strings"
)

const (
	// MaxNodes ограничивает количество доступных из корня пакетов в графе.
	MaxNodes = 1000
	// MaxDepth ограничивает глубину рекурсивного обхода.
	MaxDepth = 256
	visiting = 1
	visited  = 2
)

// Source возвращает непосредственные зависимости пакета из Maven или тестового файла.
type Source func(context.Context, string) ([]string, error)

// Edge — направленное ребро от пакета к его зависимости.
type Edge struct {
	From string
	To   string
}

// Graph хранит доступную из корня часть графа, порядок DFS и результаты анализа.
type Graph struct {
	Root     string
	Nodes    []string
	Edges    map[string][]string
	Cycles   []Edge
	Excluded []string
}

// traversal хранит состояние одного обхода; готовая вершина повторно не загружается.
type traversal struct {
	graph    Graph
	source   Source
	filter   string
	states   map[string]int
	excluded map[string]bool
}

// Build строит граф рекурсивным DFS, сохраняя рёбра циклов и общих зависимостей.
// Совпавшие с фильтром вершины, включая корень, исключаются до обращения к источнику.
func Build(ctx context.Context, root, filter string, source Source) (Graph, error) {
	if source == nil || root == "" {
		return Graph{}, fmt.Errorf("граф: нужны имя корня и источник зависимостей")
	}
	if err := ctx.Err(); err != nil {
		return Graph{}, err
	}
	t := traversal{graph: Graph{Root: root, Edges: make(map[string][]string)}, source: source,
		filter: filter, states: make(map[string]int), excluded: make(map[string]bool)}
	if !t.skip(root) {
		if err := t.visit(ctx, root, 0); err != nil {
			return Graph{}, err
		}
	}
	return t.graph, nil
}

// skip применяет чувствительный к регистру поиск подстроки к полному имени пакета.
func (t *traversal) skip(name string) bool {
	if t.filter == "" || !strings.Contains(name, t.filter) {
		return false
	}
	if !t.excluded[name] {
		t.graph.Excluded = append(t.graph.Excluded, name)
		t.excluded[name] = true
	}
	return true
}

// visit отмечает вершину до рекурсивного спуска, предотвращая бесконечный обход цикла.
func (t *traversal) visit(ctx context.Context, name string, depth int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if depth >= MaxDepth || len(t.graph.Nodes) >= MaxNodes {
		return fmt.Errorf("граф: превышен предел глубины %d или числа пакетов %d", MaxDepth, MaxNodes)
	}
	t.states[name] = visiting
	t.graph.Nodes = append(t.graph.Nodes, name)
	t.graph.Edges[name] = nil
	dependencies, err := t.source(ctx, name)
	if err != nil {
		return fmt.Errorf("зависимости %s: %w", name, err)
	}
	seen := make(map[string]bool)
	for _, dependency := range dependencies {
		if seen[dependency] || t.skip(dependency) {
			continue
		}
		seen[dependency] = true
		if err := t.follow(ctx, name, dependency, depth); err != nil {
			return err
		}
	}
	t.states[name] = visited
	return ctx.Err()
}

// follow сохраняет каждое ребро и рекурсивно обходит только ещё не посещённую вершину.
func (t *traversal) follow(ctx context.Context, from, to string, depth int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if to == "" {
		return fmt.Errorf("зависимости %s: пустое имя пакета", from)
	}
	t.graph.Edges[from] = append(t.graph.Edges[from], to)
	switch t.states[to] {
	case visiting:
		t.graph.Cycles = append(t.graph.Cycles, Edge{From: from, To: to})
	case visited:
		return nil
	default:
		return t.visit(ctx, to, depth+1)
	}
	return nil
}
