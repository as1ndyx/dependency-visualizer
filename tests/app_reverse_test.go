package tests

import (
	"strings"
	"testing"

	"dependencyvisualizer/src/app"
)

func TestRunReverseWholeTestRepository(t *testing.T) {
	path := writeTestGraphConfig(t, "A: B C\nB: D\nC: D\nD:\nX: D\nZ:\n", "A", "")
	var out, errOut strings.Builder
	code := app.Run([]string{"-config", path, "-reverse", "D"}, &out, &errOut)
	want := "Обратные зависимости D:\nОбласть поиска: весь тестовый репозиторий\n" +
		"Прямые: B, C, X\nВсе (прямые и транзитивные, DFS): B, A, C, X\n"
	if code != app.ExitOK || out.String() != want || errOut.Len() != 0 {
		t.Fatalf("код %d, stdout %q, stderr %q", code, out.String(), errOut.String())
	}
}

func TestRunReverseUsesTargetNotConfigRoot(t *testing.T) {
	path := writeTestGraphConfig(t, "A: D\nD:\nX: D\n", "D", "")
	var out, errOut strings.Builder
	code := app.Run([]string{"-reverse", "D", "-config", path}, &out, &errOut)
	if code != app.ExitOK || !strings.Contains(out.String(), "Прямые: A, X") {
		t.Fatalf("поиск ограничился зависимостями D: %q, %q", out.String(), errOut.String())
	}
}

func TestRunReverseErrors(t *testing.T) {
	path := writeTestGraphConfig(t, "A: B\nB:\n", "A", "B")
	cases := [][]string{
		{"-reverse"}, {"-reverse", ""}, {"-reverse", "  "},
		{"-config", path, "-reverse", "MISSING"}, {"-config", path, "-reverse", "B"},
		{"-config", path, "-reverse", "A", "extra"},
	}
	for _, args := range cases {
		var out, errOut strings.Builder
		code := app.Run(args, &out, &errOut)
		if code != app.ExitError || out.Len() != 0 || !strings.Contains(errOut.String(), "ошибка:") {
			t.Fatalf("аргументы %v: код %d, stdout %q, stderr %q", args, code, out.String(), errOut.String())
		}
	}
}

func TestRunReverseInvalidFileAndOutputError(t *testing.T) {
	path := writeTestGraphConfig(t, "A: MISSING", "A", "")
	var out, errOut strings.Builder
	if code := app.Run([]string{"-config", path, "-reverse", "A"}, &out, &errOut); code != app.ExitError {
		t.Fatal("принят некорректный граф")
	}
	path = writeTestGraphConfig(t, "A:\n", "A", "")
	errOut.Reset()
	code := app.Run([]string{"-config", path, "-reverse", "A"}, brokenWriter{}, &errOut)
	if code != app.ExitError || !strings.Contains(errOut.String(), "ошибка вывода") {
		t.Fatalf("ошибка вывода не обработана: %q", errOut.String())
	}
}

func TestRunReverseFilter(t *testing.T) {
	path := writeTestGraphConfig(t, "A: BX\nBX: D\nC: D\nD:\n", "A", "B")
	var out, errOut strings.Builder
	code := app.Run([]string{"-config", path, "-reverse", "D"}, &out, &errOut)
	if code != app.ExitOK || !strings.Contains(out.String(), "Прямые: C\n") ||
		!strings.Contains(out.String(), "Все (прямые и транзитивные, DFS): C\n") {
		t.Fatalf("фильтр не применён: %q, %q", out.String(), errOut.String())
	}
}

// TestRunReverseMavenScopeAndNoExtraRequests проверяет запросы только к app и lib при обратном обходе.
func TestRunReverseMavenScopeAndNoExtraRequests(t *testing.T) {
	const expectedRequestCount = 2
	files := map[string]string{
		mavenPath("app"):     pomXML("<dependencies>" + dependencyXML("lib", "1", "") + "</dependencies>"),
		mavenPath("lib"):     pomXML(""),
		mavenPath("outside"): pomXML("<dependencies>" + dependencyXML("lib", "1", "") + "</dependencies>"),
	}
	server, requests := mavenServer(t, files)
	var out, errOut strings.Builder
	path := graphMavenConfig(t, server.URL, "")
	code := app.Run([]string{"-config", path, "-reverse", "org.example:lib:1"}, &out, &errOut)
	want := "Обратные зависимости org.example:lib:1:\nОбласть поиска: граф org.example:app:1\n" +
		"Прямые: org.example:app:1\nВсе (прямые и транзитивные, DFS): org.example:app:1\n"
	if code != app.ExitOK || out.String() != want || errOut.Len() != 0 {
		t.Fatalf("код %d, stdout %q, stderr %q", code, out.String(), errOut.String())
	}
	if len(requests()) != expectedRequestCount ||
		requests()[mavenPath("app")] != 1 || requests()[mavenPath("lib")] != 1 {
		t.Fatalf("лишние запросы при обратном обходе: %v", requests())
	}
}

func TestRunReverseMavenUnknownAndFiltered(t *testing.T) {
	for _, filter := range []string{"", "outside"} {
		server, requests := mavenServer(t, map[string]string{mavenPath("app"): pomXML("")})
		path := graphMavenConfig(t, server.URL, filter)
		var out, errOut strings.Builder
		code := app.Run([]string{"-config", path, "-reverse", "org.example:outside:1"}, &out, &errOut)
		want := "не найден в области поиска"
		if filter != "" {
			want = "исключён фильтром"
		}
		if code != app.ExitError || out.Len() != 0 || !strings.Contains(errOut.String(), want) {
			t.Fatalf("код %d, stdout %q, stderr %q", code, out.String(), errOut.String())
		}
		if requests()[mavenPath("outside")] != 0 {
			t.Fatal("запрошен пакет вне выбранного графа")
		}
		if filter != "" && len(requests()) != 0 {
			t.Fatal("были запросы при исключённой цели")
		}
	}
}
