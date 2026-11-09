package graph

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
)

// ReverseResult — прямые и все обратные зависимости выбранного пакета.
// Сам выбранный пакет не включается в списки даже при наличии циклов.
type ReverseResult struct {
	Target   string
	Direct   []string
	All      []string
	Excluded []string
}

// Reverse меняет направление рёбер и вызывает рекурсивный DFS третьего этапа.
// Область поиска задаётся переданным графом: все пакеты файла или загруженный граф Maven.
func Reverse(ctx context.Context, edges map[string][]string, target, filter string) (ReverseResult, error) {
	if err := ctx.Err(); err != nil {
		return ReverseResult{}, err
	}
	if _, exists := edges[target]; !exists {
		return ReverseResult{}, fmt.Errorf("reverse: пакет %s не найден в области поиска", target)
	}
	if filter != "" && strings.Contains(target, filter) {
		return ReverseResult{}, fmt.Errorf("reverse: пакет %s исключён фильтром", target)
	}
	incoming, err := transpose(edges)
	if err != nil {
		return ReverseResult{}, err
	}
	source := func(ctx context.Context, name string) ([]string, error) { return incoming[name], nil }
	result, err := Build(ctx, target, filter, source)
	if err != nil {
		return ReverseResult{}, err
	}
	return ReverseResult{Target: target, Direct: withoutTarget(result.Edges[target], target),
		All: withoutTarget(result.Nodes, target), Excluded: result.Excluded}, nil
}

// transpose строит обратные рёбра в стабильном порядке и проверяет ссылки на вершины.
func transpose(edges map[string][]string) (map[string][]string, error) {
	incoming := make(map[string][]string, len(edges))
	names := make([]string, 0, len(edges))
	for name := range edges {
		incoming[name] = nil
		names = append(names, name)
	}
	sort.Strings(names)
	for _, from := range names {
		for _, to := range edges[from] {
			if _, exists := incoming[to]; !exists {
				return nil, fmt.Errorf("reverse: %s ссылается на отсутствующую вершину %s", from, to)
			}
			incoming[to] = append(incoming[to], from)
		}
	}
	return incoming, nil
}

// withoutTarget удаляет сам анализируемый пакет, сохраняя порядок остальных вершин.
func withoutTarget(names []string, target string) []string {
	var result []string
	for _, name := range names {
		if name != target {
			result = append(result, name)
		}
	}
	return result
}

// PrintReverse выводит область поиска и списки прямых и транзитивных обратных зависимостей.
func PrintReverse(out io.Writer, result ReverseResult, scope string) error {
	var text strings.Builder
	fmt.Fprintf(&text, "Обратные зависимости %s:\nОбласть поиска: %s\n", result.Target, scope)
	fmt.Fprintf(&text, "Прямые: %s\n", listOrNone(result.Direct))
	fmt.Fprintf(&text, "Все (прямые и транзитивные, DFS): %s\n", listOrNone(result.All))
	if len(result.Excluded) != 0 {
		fmt.Fprintf(&text, "Исключены фильтром: %s\n", strings.Join(result.Excluded, ", "))
	}
	_, err := io.WriteString(out, text.String())
	return err
}

// listOrNone обозначает отсутствие зависимых пакетов вместо пустой строки.
func listOrNone(names []string) string {
	if len(names) == 0 {
		return "(нет)"
	}
	return strings.Join(names, ", ")
}
