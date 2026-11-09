package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dependencyvisualizer/src/config"
)

const validINI = `[settings]
package = org.apache.commons:commons-lang3:3.12.0
repository = https://repo.maven.apache.org/maven2
test_mode = false
filter = junit
`

// writeConfig создаёт независимый временный INI-файл для каждого теста.
func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.ini")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadMaven(t *testing.T) {
	path := writeConfig(t, "\uFEFF; комментарий\r\n# ещё один\r\n"+strings.ReplaceAll(validINI, "\n", "\r\n"))
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := config.Config{
		Package:    "org.apache.commons:commons-lang3:3.12.0",
		Repository: "https://repo.maven.apache.org/maven2", Filter: "junit",
	}
	if cfg != want {
		t.Fatalf("получили %+v, ожидали %+v", cfg, want)
	}
}

func TestLoadRequiredParameters(t *testing.T) {
	for _, key := range []string{"package", "repository", "test_mode", "filter"} {
		t.Run(key, func(t *testing.T) {
			var lines []string
			for _, line := range strings.Split(validINI, "\n") {
				if !strings.HasPrefix(line, key+" =") {
					lines = append(lines, line)
				}
			}
			_, err := config.Load(writeConfig(t, strings.Join(lines, "\n")))
			if err == nil || !strings.Contains(err.Error(), key) {
				t.Fatalf("ожидали ошибку параметра %s, получили %v", key, err)
			}
		})
	}
}

func TestLoadInvalidValues(t *testing.T) {
	cases := []struct{ key, value string }{
		{"package", ""}, {"package", "commons-lang3"}, {"package", "org:lib:"},
		{"package", "org:../lib:1.0"}, {"repository", ""}, {"repository", "relative/path"},
		{"repository", "ftp://example.org"}, {"repository", "https://"},
		{"repository", "https://host:bad"}, {"repository", "https://host/%zz"},
		{"repository", "https://host/#fragment"}, {"repository", "https://host/?query=1"},
		{"repository", "https://user:password@host/"}, {"test_mode", ""},
		{"test_mode", "yes"}, {"test_mode", "1"}, {"filter", "bad\x00value"},
		{"filter", "bad\tvalue"}, {"filter", "bad\xffvalue"},
	}
	for _, tc := range cases {
		t.Run(tc.key+"/"+tc.value, func(t *testing.T) {
			lines := strings.Split(validINI, "\n")
			for i, line := range lines {
				if strings.HasPrefix(line, tc.key+" =") {
					lines[i] = tc.key + " = " + tc.value
				}
			}
			_, err := config.Load(writeConfig(t, strings.Join(lines, "\n")))
			if err == nil || !strings.Contains(err.Error(), tc.key) {
				t.Fatalf("ожидали ошибку параметра %s, получили %v", tc.key, err)
			}
		})
	}
}

func TestLoadINIErrors(t *testing.T) {
	cases := []struct{ name, text, want string }{
		{"empty", "", "[settings]"},
		{"section", strings.ReplaceAll(validINI, "[settings]", "[other]"), "неизвестная секция"},
		{"outside", strings.TrimPrefix(validINI, "[settings]\n"), "вне секции"},
		{"duplicate-section", validINI + "[settings]\n", "повторно"},
		{"duplicate-key", validINI + "filter = x\n", "filter: параметр указан повторно"},
		{"unknown", validINI + "typo = value\n", "неизвестный параметр"},
		{"syntax", validINI + "broken\n", "ключ = значение"},
		{"long", validINI + strings.Repeat("a", 100000), "ошибка чтения"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := config.Load(writeConfig(t, tc.text))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ожидали %q, получили %v", tc.want, err)
			}
		})
	}
}

func TestLoadLiteralFilter(t *testing.T) {
	for _, value := range []string{"", "тест", "junit;test#x=y"} {
		t.Run(value, func(t *testing.T) {
			cfg, err := config.Load(writeConfig(t, strings.Replace(validINI, "junit", value, 1)))
			if err != nil || cfg.Filter != value {
				t.Fatalf("filter: получили %q, ошибка %v", cfg.Filter, err)
			}
		})
	}
}

func TestLoadConfigFileErrors(t *testing.T) {
	for _, path := range []string{filepath.Join(t.TempDir(), "missing.ini"), t.TempDir()} {
		if _, err := config.Load(path); err == nil {
			t.Fatalf("ожидали ошибку для %s", path)
		}
	}
}
