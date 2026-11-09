package tests

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"dependencyvisualizer/src/app"
)

func TestRunPrintsAllParameters(t *testing.T) {
	path := writeConfig(t, validINI)
	var out, errOut strings.Builder
	code := app.Run([]string{"-config", path}, &out, &errOut)
	want := "package = org.apache.commons:commons-lang3:3.12.0\n" +
		"repository = https://repo.maven.apache.org/maven2\ntest_mode = false\nfilter = junit\n"
	if code != app.ExitOK || out.String() != want || errOut.Len() != 0 {
		t.Fatalf("код %d, stdout %q, stderr %q", code, out.String(), errOut.String())
	}
}

func TestRunDefaultConfig(t *testing.T) {
	path := writeConfig(t, validINI)
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
	code := app.Run([]string{"-config", writeConfig(t, validINI)}, brokenWriter{}, &errOut)
	if code != app.ExitError || !strings.Contains(errOut.String(), "ошибка вывода") {
		t.Fatalf("код %d, stderr %q", code, errOut.String())
	}
}
