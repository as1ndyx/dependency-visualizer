package tests

import (
	"strings"
	"testing"

	"dependencyvisualizer/src/app"
	"dependencyvisualizer/src/graph"
)

// renderDOT возвращает описание для проверки представления графа этапа 5.
func renderDOT(t *testing.T, value graph.Graph) string {
	t.Helper()
	var out strings.Builder
	if err := graph.PrintDOT(&out, value); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

func TestDOTSharedDependencyAndCycle(t *testing.T) {
	const expectedEdgeCount = 5 // Четыре ребра ромба и замыкающее цикл ребро D -> A.
	repository := map[string][]string{"A": {"B", "C"}, "B": {"D"}, "C": {"D"}, "D": {"A"}}
	value := buildTestGraph(t, repository, "A", "")
	got := renderDOT(t, value)
	for _, edge := range []string{`"A" -> "B";`, `"A" -> "C";`, `"B" -> "D";`, `"C" -> "D";`} {
		if strings.Count(got, edge) != 1 {
			t.Errorf("потеряно или повторено ребро %s: %s", edge, got)
		}
	}
	if !strings.Contains(got, `"D" -> "A" [color="#dc2626", penwidth=2];`) {
		t.Fatal("ребро цикла не выделено")
	}
	if strings.Count(got, `color="#dc2626"`) != 1 || strings.Count(got, " -> ") != expectedEdgeCount {
		t.Fatal("ложный цикл или лишние рёбра")
	}
	if renderDOT(t, value) != got {
		t.Fatal("нестабильный порядок вывода")
	}
}

func TestDOTIsolatedAndFilteredRoot(t *testing.T) {
	for _, filter := range []string{"", "A"} {
		value := buildTestGraph(t, map[string][]string{"A": {}}, "A", filter)
		got := renderDOT(t, value)
		if !strings.HasPrefix(got, "digraph dependencies {\n") || !strings.HasSuffix(got, "}\n") {
			t.Fatalf("неверные границы DOT: %s", got)
		}
		if filter == "" && !strings.Contains(got, `"A" [label="A", fillcolor="#dbeafe"`) {
			t.Fatal("одиночный корень отсутствует или не выделен")
		}
		if filter != "" && (strings.Contains(got, `"A" [`) || !strings.Contains(got, "Исключены: A")) {
			t.Fatal("исключённый корень остался в графе или отсутствует пояснение")
		}
	}
}

func TestDOTQuotesAndMavenLabels(t *testing.T) {
	name := "пакет\"\\\n\r\t"
	value := graph.Graph{Root: name, Nodes: []string{name, "org.example:lib:1"},
		Edges: map[string][]string{name: {"org.example:lib:1"}}}
	got := renderDOT(t, value)
	if !strings.Contains(got, `"пакет\"\\\n\r\t" -> "org.example:lib:1";`) {
		t.Fatalf("не экранированы специальные символы: %s", got)
	}
	if !strings.Contains(got, `"org.example:lib:1" [label="org.example\nlib\n1"]`) {
		t.Fatal("координаты листа должны сохранять ID и отображаться на трёх строках")
	}
	if err := graph.PrintDOT(brokenWriter{}, value); err == nil {
		t.Fatal("потеряна ошибка записи")
	}
}

func TestRunDefaultDOT(t *testing.T) {
	path := writeTestGraphConfig(t, "A: B\nB:\n", "A", "")
	var out, errOut strings.Builder
	code := app.Run([]string{"-config", path}, &out, &errOut)
	if code != app.ExitOK || errOut.Len() != 0 || !strings.HasPrefix(out.String(), "digraph dependencies {\n") {
		t.Fatalf("код %d, stdout %q, stderr %q", code, out.String(), errOut.String())
	}
	if !strings.Contains(out.String(), `"A" -> "B";`) || !strings.HasSuffix(out.String(), "}\n") {
		t.Fatal("stdout должен содержать полное описание диаграммы")
	}
	code = app.Run([]string{"-config", path, "-format", "dot"}, brokenWriter{}, &errOut)
	if code != app.ExitError || !strings.Contains(errOut.String(), "ошибка вывода") {
		t.Fatal("CLI не обработал ошибку записи DOT")
	}
}

func TestRunInvalidFormat(t *testing.T) {
	cases := [][]string{
		{"-format", "svg"}, {"-format", ""}, {"-format"},
		{"-reverse", "A", "-format", "dot"},
	}
	for _, args := range cases {
		var out, errOut strings.Builder
		if app.Run(args, &out, &errOut) != app.ExitError || out.Len() != 0 || errOut.Len() == 0 {
			t.Fatalf("неверный формат принят: %v", args)
		}
	}
}
