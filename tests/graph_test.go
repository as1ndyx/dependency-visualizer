package tests

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"dependencyvisualizer/src/graph"
)

// buildTestGraph проверяет, что каждый включённый пакет запрашивается ровно один раз.
func buildTestGraph(t *testing.T, repository map[string][]string, root, filter string) graph.Graph {
	t.Helper()
	calls := make(map[string]int)
	source := func(ctx context.Context, name string) ([]string, error) {
		calls[name]++
		if filter != "" && strings.Contains(name, filter) {
			t.Errorf("исключённый пакет %s был загружен", name)
		}
		return repository[name], nil
	}
	result, err := graph.Build(context.Background(), root, filter, source)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range result.Nodes {
		if calls[node] != 1 {
			t.Errorf("пакет %s запрошен %d раз", node, calls[node])
		}
	}
	return result
}

func TestGraphDFSAndSharedDependency(t *testing.T) {
	repository := map[string][]string{"A": {"B", "C", "B"}, "B": {"D"}, "C": {"D"}, "D": {}, "X": {}}
	got := buildTestGraph(t, repository, "A", "")
	if !reflect.DeepEqual(got.Nodes, []string{"A", "B", "D", "C"}) {
		t.Fatalf("нарушен порядок DFS или добавлена недостижимая вершина: %v", got.Nodes)
	}
	want := map[string][]string{"A": {"B", "C"}, "B": {"D"}, "C": {"D"}, "D": nil}
	if !reflect.DeepEqual(got.Edges, want) || len(got.Cycles) != 0 {
		t.Fatalf("потеряны рёбра ромба или обнаружен ложный цикл: %+v", got)
	}
}

func TestGraphCycles(t *testing.T) {
	cases := []struct {
		name       string
		repository map[string][]string
		cycles     []graph.Edge
	}{
		{"cycle", map[string][]string{"A": {"B"}, "B": {"C"}, "C": {"A"}}, []graph.Edge{{From: "C", To: "A"}}},
		{"self", map[string][]string{"A": {"A", "A"}}, []graph.Edge{{From: "A", To: "A"}}},
		{"two-cycles", map[string][]string{"A": {"B", "C"}, "B": {"A"}, "C": {"A"}},
			[]graph.Edge{{From: "B", To: "A"}, {From: "C", To: "A"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildTestGraph(t, tc.repository, "A", "")
			if !reflect.DeepEqual(got.Cycles, tc.cycles) {
				t.Fatalf("ожидали %v, получили %v", tc.cycles, got.Cycles)
			}
			for _, edge := range tc.cycles {
				if !strings.Contains(strings.Join(got.Edges[edge.From], " "), edge.To) {
					t.Errorf("ребро цикла удалено: %v", edge)
				}
			}
		})
	}
}

func TestGraphFilterPrunesBranch(t *testing.T) {
	repository := map[string][]string{"A": {"BX", "C"}, "BX": {"D"}, "C": {}, "D": {}}
	got := buildTestGraph(t, repository, "A", "B")
	if !reflect.DeepEqual(got.Nodes, []string{"A", "C"}) || !reflect.DeepEqual(got.Edges["A"], []string{"C"}) {
		t.Fatalf("исключённая ветвь осталась в графе: %+v", got)
	}
	if !reflect.DeepEqual(got.Excluded, []string{"BX"}) {
		t.Fatalf("неверный список исключений: %v", got.Excluded)
	}
	repository["C"] = []string{"D", "BX"}
	got = buildTestGraph(t, repository, "A", "B")
	if !reflect.DeepEqual(got.Nodes, []string{"A", "C", "D"}) || len(got.Excluded) != 1 {
		t.Fatalf("потерян альтернативный путь или дублируется исключение: %+v", got)
	}
}

func TestGraphRootFilterAndCase(t *testing.T) {
	repository := map[string][]string{"ROOT": {"B"}, "B": {}}
	got := buildTestGraph(t, repository, "ROOT", "OO")
	if len(got.Nodes) != 0 || !reflect.DeepEqual(got.Excluded, []string{"ROOT"}) {
		t.Fatalf("корень не исключён: %+v", got)
	}
	got = buildTestGraph(t, repository, "ROOT", "oo")
	if len(got.Nodes) != 2 {
		t.Fatalf("поиск подстроки должен учитывать регистр: %+v", got)
	}
}

func TestGraphSourceError(t *testing.T) {
	source := func(ctx context.Context, name string) ([]string, error) {
		if name == "A" {
			return []string{"B"}, nil
		}
		return nil, fmt.Errorf("POM не найден")
	}
	got, err := graph.Build(context.Background(), "A", "", source)
	if err == nil || !strings.Contains(err.Error(), "зависимости B") || len(got.Nodes) != 0 {
		t.Fatalf("ожидали ошибку без частичного графа: %+v, %v", got, err)
	}
}
