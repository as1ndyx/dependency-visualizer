package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dependencyvisualizer/src/app"
)

// writeTestGraphConfig сохраняет граф рядом с INI, чтобы проверить относительные пути.
func writeTestGraphConfig(t *testing.T, graph, root, filter string) string {
	t.Helper()
	text := strings.Replace(testINI, "package = A", "package = "+root, 1)
	text = strings.Replace(text, "filter =", "filter = "+filter, 1)
	path := writeConfig(t, text)
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "graph.txt"), []byte(graph), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunTestGraphRelativeAndAbsolutePath(t *testing.T) {
	path := writeTestGraphConfig(t, "A: B\nB: A C\nC:\n", "A", "C")
	for _, absolute := range []bool{false, true} {
		if absolute {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			text := strings.Replace(string(data), "graph.txt", filepath.Join(filepath.Dir(path), "graph.txt"), 1)
			if err := os.WriteFile(path, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
		}
		var out, errOut strings.Builder
		code := app.Run([]string{"-config", path}, &out, &errOut)
		want := "Граф зависимостей A (DFS):\nA -> B\nB -> A\n" +
			"Циклы (обратные рёбра DFS):\nB -> A\nИсключены фильтром: C\n"
		if code != app.ExitOK || out.String() != want || errOut.Len() != 0 {
			t.Fatalf("код %d, stdout %q, stderr %q", code, out.String(), errOut.String())
		}
	}
}

func TestRunTestGraphErrors(t *testing.T) {
	for _, text := range []string{"B:", "A: MISSING", "A B", "A:\nA:"} {
		path := writeTestGraphConfig(t, text, "A", "")
		var out, errOut strings.Builder
		code := app.Run([]string{"-config", path}, &out, &errOut)
		if code != app.ExitError || out.Len() != 0 || !strings.Contains(errOut.String(), "ошибка:") {
			t.Fatalf("код %d, stdout %q, stderr %q", code, out.String(), errOut.String())
		}
	}
}

// graphMavenConfig задаёт корень и фильтр для интеграционных проверок HTTP-графа.
func graphMavenConfig(t *testing.T, repository, filter string) string {
	t.Helper()
	text := strings.Replace(validINI, "org.apache.commons:commons-lang3:3.12.0", "org.example:app:1", 1)
	text = strings.Replace(text, "https://repo.maven.apache.org/maven2", repository, 1)
	text = strings.Replace(text, "filter = junit", "filter = "+filter, 1)
	return writeConfig(t, text)
}

func TestRunMavenTransitiveCycle(t *testing.T) {
	files := map[string]string{
		mavenPath("parent"): pomXML(""),
		mavenPath("app"): pomXML("<dependencies>" + dependencyXML("b", "1", "") +
			dependencyXML("c", "1", "") + "</dependencies>"),
		mavenPath("b"): pomXML(parentXML("parent") + "<dependencies>" + dependencyXML("d", "1", "") + "</dependencies>"),
		mavenPath("c"): pomXML(parentXML("parent") + "<dependencies>" + dependencyXML("d", "1", "") + "</dependencies>"),
		mavenPath("d"): pomXML("<dependencies>" + dependencyXML("app", "1", "") + "</dependencies>"),
	}
	server, requests := mavenServer(t, files)
	var out, errOut strings.Builder
	code := app.Run([]string{"-config", graphMavenConfig(t, server.URL, "")}, &out, &errOut)
	want := "Граф зависимостей org.example:app:1 (DFS):\n" +
		"org.example:app:1 -> org.example:b:1, org.example:c:1\n" +
		"org.example:b:1 -> org.example:d:1\norg.example:d:1 -> org.example:app:1\n" +
		"org.example:c:1 -> org.example:d:1\nЦиклы (обратные рёбра DFS):\norg.example:d:1 -> org.example:app:1\n"
	if code != app.ExitOK || out.String() != want || errOut.Len() != 0 {
		t.Fatalf("код %d, stdout %q, stderr %q", code, out.String(), errOut.String())
	}
	if len(requests()) != len(files) {
		t.Fatalf("не все пакеты загружены: %v", requests())
	}
	for path, count := range requests() {
		if count != 1 {
			t.Errorf("POM %s запрошен %d раз", path, count)
		}
	}
}

func TestRunMavenFilterBeforeDownload(t *testing.T) {
	for _, filter := range []string{"skip", "app"} {
		body := pomXML("<dependencies>" + dependencyXML("skip", "1", "") + "</dependencies>")
		server, requests := mavenServer(t, map[string]string{mavenPath("app"): body})
		var out, errOut strings.Builder
		code := app.Run([]string{"-config", graphMavenConfig(t, server.URL, filter)}, &out, &errOut)
		if code != app.ExitOK || errOut.Len() != 0 || !strings.Contains(out.String(), "Исключены фильтром") {
			t.Fatalf("код %d, stdout %q, stderr %q", code, out.String(), errOut.String())
		}
		if requests()[mavenPath("skip")] != 0 {
			t.Fatal("загружена исключённая зависимость")
		}
		if filter == "app" && len(requests()) != 0 {
			t.Fatal("загружен исключённый корень")
		}
	}
}

func TestRunMavenTransitiveError(t *testing.T) {
	body := pomXML("<dependencies>" + dependencyXML("missing", "1", "") + "</dependencies>")
	server, _ := mavenServer(t, map[string]string{mavenPath("app"): body})
	var out, errOut strings.Builder
	code := app.Run([]string{"-config", graphMavenConfig(t, server.URL, "")}, &out, &errOut)
	if code != app.ExitError || out.Len() != 0 || !strings.Contains(errOut.String(), "HTTP 404") ||
		!strings.Contains(errOut.String(), "зависимости org.example:missing:1") {
		t.Fatalf("код %d, stdout %q, stderr %q", code, out.String(), errOut.String())
	}
}
