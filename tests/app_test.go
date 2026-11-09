package tests

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"dependencyvisualizer/src/app"
)

func TestRunPrintsDirectDependencies(t *testing.T) {
	path := writeMavenConfig(t, "<dependencies>"+dependencyXML("lib", "2", "")+"</dependencies>")
	var out, errOut strings.Builder
	code := app.Run([]string{"-config", path}, &out, &errOut)
	want := "Прямые зависимости org.example:app:1:\norg.example:lib:2 [scope=compile]\n"
	if code != app.ExitOK || out.String() != want || errOut.Len() != 0 {
		t.Fatalf("код %d, stdout %q, stderr %q", code, out.String(), errOut.String())
	}
}

func TestRunDefaultConfig(t *testing.T) {
	path := writeMavenConfig(t, "<dependencies>"+dependencyXML("lib", "2", "")+"</dependencies>")
	t.Chdir(filepath.Dir(path))
	var out, errOut strings.Builder
	if code := app.Run(nil, &out, &errOut); code != app.ExitOK || out.Len() == 0 {
		t.Fatalf("код %d, stderr %q", code, errOut.String())
	}
}

func TestRunErrors(t *testing.T) {
	bad := writeConfig(t, strings.Replace(validINI, "test_mode = false", "test_mode = yes", 1))
	cases := [][]string{
		{"-unknown"}, {"-config"}, {"-config", ""}, {"extra"},
		{"-config", bad}, {"-config", "missing.ini"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var out, errOut strings.Builder
			code := app.Run(args, &out, &errOut)
			if code != app.ExitError || out.Len() != 0 || !strings.Contains(errOut.String(), "ошибка:") {
				t.Fatalf("код %d, stdout %q, stderr %q", code, out.String(), errOut.String())
			}
		})
	}
}

func TestRunHelp(t *testing.T) {
	var out, errOut strings.Builder
	if code := app.Run([]string{"-help"}, &out, &errOut); code != app.ExitOK {
		t.Fatalf("справка завершилась с кодом %d", code)
	}
	if !strings.Contains(errOut.String(), "-config") || out.Len() != 0 {
		t.Fatalf("неверный вывод справки: %q", errOut.String())
	}
}

type brokenWriter struct{}

func (brokenWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("канал вывода закрыт")
}

func TestRunOutputError(t *testing.T) {
	var errOut strings.Builder
	code := app.Run([]string{"-config", writeMavenConfig(t, "")}, brokenWriter{}, &errOut)
	if code != app.ExitError || !strings.Contains(errOut.String(), "ошибка вывода") {
		t.Fatalf("код %d, stderr %q", code, errOut.String())
	}
}

func TestRunTestModeNotImplemented(t *testing.T) {
	var out, errOut strings.Builder
	code := app.Run([]string{"-config", "../examples/test.ini"}, &out, &errOut)
	if code != app.ExitError || out.Len() != 0 || !strings.Contains(errOut.String(), "этапе 3") {
		t.Fatalf("код %d, stdout %q, stderr %q", code, out.String(), errOut.String())
	}
}

func TestRunMavenError(t *testing.T) {
	path := writeMavenConfig(t, "<dependencies>"+dependencyXML("lib", "", "")+"</dependencies>")
	var out, errOut strings.Builder
	code := app.Run([]string{"-config", path}, &out, &errOut)
	if code != app.ExitError || out.Len() != 0 || !strings.Contains(errOut.String(), "версию") {
		t.Fatalf("код %d, stdout %q, stderr %q", code, out.String(), errOut.String())
	}
}
