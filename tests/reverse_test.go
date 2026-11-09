package tests

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"dependencyvisualizer/src/graph"
)

func TestReverseChainAndDiamond(t *testing.T) {
	cases := []struct {
		name        string
		edges       map[string][]string
		direct, all []string
	}{
		{"chain", map[string][]string{"A": {"B"}, "B": {"D"}, "D": nil}, []string{"B"}, []string{"B", "A"}},
		{"diamond", map[string][]string{"A": {"B", "C"}, "B": {"D"}, "C": {"D"}, "D": nil},
			[]string{"B", "C"}, []string{"B", "A", "C"}},
		{"disconnected", map[string][]string{"A": {"D"}, "D": nil, "X": {"D"}, "Z": nil},
			[]string{"A", "X"}, []string{"A", "X"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := graph.Reverse(context.Background(), tc.edges, "D", "")
			if err != nil || !reflect.DeepEqual(got.Direct, tc.direct) || !reflect.DeepEqual(got.All, tc.all) {
				t.Fatalf("получили %+v, ошибка %v", got, err)
			}
		})
	}
}

func TestReverseCyclesAndDuplicates(t *testing.T) {
	edges := map[string][]string{"A": {"B", "B", "A"}, "B": {"C"}, "C": {"A"}}
	got, err := graph.Reverse(context.Background(), edges, "A", "")
	if err != nil || !reflect.DeepEqual(got.Direct, []string{"C"}) || !reflect.DeepEqual(got.All, []string{"C", "B"}) {
		t.Fatalf("цикл обработан неверно: %+v, %v", got, err)
	}
	if !reflect.DeepEqual(edges["A"], []string{"B", "B", "A"}) {
		t.Fatal("обратный анализ изменил исходный граф")
	}
}

func TestReverseNoDependents(t *testing.T) {
	for _, edges := range []map[string][]string{{"A": nil}, {"A": {"A"}}, {"A": {"B"}, "B": nil}} {
		got, err := graph.Reverse(context.Background(), edges, "A", "")
		if err != nil || len(got.Direct) != 0 || len(got.All) != 0 {
			t.Fatalf("сам пакет или его исходящая зависимость попали в результат: %+v, %v", got, err)
		}
	}
}

func TestReverseFilter(t *testing.T) {
	edges := map[string][]string{"A": {"BX"}, "BX": {"D"}, "C": {"D"}, "D": nil}
	got, err := graph.Reverse(context.Background(), edges, "D", "B")
	if err != nil || !reflect.DeepEqual(got.All, []string{"C"}) || !reflect.DeepEqual(got.Excluded, []string{"BX"}) {
		t.Fatalf("фильтр не отсекает обратную ветвь: %+v, %v", got, err)
	}
	edges["A"] = []string{"BX", "C"}
	got, err = graph.Reverse(context.Background(), edges, "D", "B")
	if err != nil || !reflect.DeepEqual(got.All, []string{"C", "A"}) {
		t.Fatalf("потерян альтернативный путь: %+v, %v", got, err)
	}
}

func TestReverseErrors(t *testing.T) {
	cases := []struct {
		edges                map[string][]string
		target, filter, want string
	}{
		{map[string][]string{"A": nil}, "B", "", "не найден"},
		{map[string][]string{"A": nil}, "A", "A", "исключён фильтром"},
		{map[string][]string{"A": {"B"}}, "A", "", "отсутствующую вершину"},
	}
	for _, tc := range cases {
		_, err := graph.Reverse(context.Background(), tc.edges, tc.target, tc.filter)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("ожидали %q, получили %v", tc.want, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := graph.Reverse(ctx, map[string][]string{"A": nil}, "A", ""); err != context.Canceled {
		t.Fatalf("отмена не обработана: %v", err)
	}
}

func TestReverseOutput(t *testing.T) {
	result := graph.ReverseResult{Target: "D", Direct: []string{"B"}, All: []string{"B", "A"}, Excluded: []string{"C"}}
	var out strings.Builder
	if err := graph.PrintReverse(&out, result, "весь тестовый репозиторий"); err != nil {
		t.Fatal(err)
	}
	want := "Обратные зависимости D:\nОбласть поиска: весь тестовый репозиторий\n" +
		"Прямые: B\nВсе (прямые и транзитивные, DFS): B, A\nИсключены фильтром: C\n"
	if out.String() != want {
		t.Fatalf("неверный вывод: %q", out.String())
	}
	out.Reset()
	if err := graph.PrintReverse(&out, graph.ReverseResult{Target: "A"}, "граф A"); err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.String(), "(нет)") != 2 {
		t.Fatal(out.String())
	}
	if err := graph.PrintReverse(brokenWriter{}, result, "граф A"); err == nil {
		t.Fatal("потеряна ошибка вывода")
	}
}
