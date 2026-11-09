package tests

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"dependencyvisualizer/src/graph"
)

func TestGraphLimits(t *testing.T) {
	chain := func(ctx context.Context, name string) ([]string, error) {
		index, _ := strconv.Atoi(name)
		return []string{strconv.Itoa(index + 1)}, nil
	}
	star := func(ctx context.Context, name string) ([]string, error) {
		if name != "0" {
			return nil, nil
		}
		var names []string
		for i := 1; i <= graph.MaxNodes; i++ {
			names = append(names, strconv.Itoa(i))
		}
		return names, nil
	}
	for _, source := range []graph.Source{chain, star} {
		_, err := graph.Build(context.Background(), "0", "", source)
		if err == nil || !strings.Contains(err.Error(), "предел") {
			t.Fatalf("нет ограничения: %v", err)
		}
	}
}

func TestGraphCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	source := func(context.Context, string) ([]string, error) {
		t.Fatal("источник вызван после отмены")
		return nil, nil
	}
	if _, err := graph.Build(ctx, "A", "", source); err != context.Canceled {
		t.Fatalf("отмена: %v", err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	source = func(context.Context, string) ([]string, error) { cancel(); return nil, nil }
	if _, err := graph.Build(ctx, "A", "", source); err != context.Canceled {
		t.Fatalf("отмена в источнике: %v", err)
	}
}

func TestGraphInvalidInput(t *testing.T) {
	if _, err := graph.Build(context.Background(), "A", "", nil); err == nil {
		t.Fatal("принят пустой источник")
	}
	source := func(context.Context, string) ([]string, error) { return []string{""}, nil }
	for _, root := range []string{"", "A"} {
		if _, err := graph.Build(context.Background(), root, "", source); err == nil {
			t.Fatal("принято пустое имя")
		}
	}
}

func TestGraphPrint(t *testing.T) {
	got := buildTestGraph(t, map[string][]string{"A": {"B", "C"}, "B": {"A"}, "C": {}}, "A", "C")
	var out strings.Builder
	if err := graph.Print(&out, got); err != nil {
		t.Fatal(err)
	}
	want := "Граф зависимостей A (DFS):\nA -> B\nB -> A\n" +
		"Циклы (обратные рёбра DFS):\nB -> A\nИсключены фильтром: C\n"
	if out.String() != want {
		t.Fatalf("неверный вывод:\n%s", out.String())
	}
	if err := graph.Print(brokenWriter{}, got); err == nil {
		t.Fatal("потеряна ошибка вывода")
	}
}

func TestGraphPrintExcludedRoot(t *testing.T) {
	got := buildTestGraph(t, nil, "ROOT", "ROOT")
	var out strings.Builder
	if err := graph.Print(&out, got); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "корневой пакет исключён") {
		t.Fatal(fmt.Sprint(out.String()))
	}
}
