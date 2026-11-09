package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dependencyvisualizer/src/config"
)

const testINI = `[settings]
package = A
repository = graph.txt
test_mode = true
filter =
`

func TestLoadTestRepository(t *testing.T) {
	path := writeConfig(t, testINI)
	repository := filepath.Join(filepath.Dir(path), "graph.txt")
	if err := os.WriteFile(repository, []byte("A: B\nB:\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"graph.txt", repository} {
		if err := os.WriteFile(path, []byte(strings.Replace(testINI, "graph.txt", value, 1)), 0600); err != nil {
			t.Fatal(err)
		}
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Package != "A" || cfg.Repository != value || !cfg.TestMode || cfg.Filter != "" {
			t.Fatalf("неверная конфигурация: %+v", cfg)
		}
	}
}

func TestLoadTestRepositoryErrors(t *testing.T) {
	cases := []struct{ old, value, want string }{
		{"graph.txt", "", "repository"},
		{"graph.txt", "missing.txt", "repository"},
		{"graph.txt", ".", "обычный файл"},
		{"package = A", "package = a", "package"},
		{"package = A", "package = А", "package"},
		{"package = A", "package = A1", "package"},
	}
	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			_, err := config.Load(writeConfig(t, strings.Replace(testINI, tc.old, tc.value, 1)))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ожидали %q, получили %v", tc.want, err)
			}
		})
	}
}
