package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
)

// Форматы имён: координаты Maven и тестовые пакеты из заглавных латинских букв.
var (
	mavenName = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]*:` +
		`[A-Za-z0-9_][A-Za-z0-9_.-]*:[A-Za-z0-9_][A-Za-z0-9_.+-]*$`)
	testName = regexp.MustCompile(`^[A-Z]+$`)
)

// validate проверяет имя пакета и источник репозитория с учётом режима.
// Filter может быть любой печатной строкой, включая пустую.
func (c Config) validate(configPath string) error {
	if c.TestMode {
		if !testName.MatchString(c.Package) {
			return fmt.Errorf("package: в тестовом режиме ожидаются заглавные латинские буквы, например A")
		}
		return validateFile(c.Repository, configPath)
	}
	if !mavenName.MatchString(c.Package) {
		return fmt.Errorf("package: ожидаются координаты Maven groupId:artifactId:version")
	}
	return validateURL(c.Repository)
}

// validateURL принимает абсолютный HTTP(S)-адрес репозитория.
// На этапе 1 проверяется синтаксис, сетевые запросы не выполняются.
func validateURL(value string) error {
	u, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("repository: некорректный URL: %w", err)
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.Opaque != "" {
		return fmt.Errorf("repository: ожидается абсолютный URL с http:// или https:// и именем хоста")
	}
	if u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return fmt.Errorf("repository: URL не должен содержать логин, параметры запроса или фрагмент")
	}
	return nil
}

// validateFile проверяет, что тестовый репозиторий является доступным обычным файлом.
// Чтение описания графа будет добавлено на этапе 3.
func validateFile(value, configPath string) error {
	if value == "" {
		return fmt.Errorf("repository: путь к тестовому репозиторию не должен быть пустым")
	}
	path := value
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(configPath), path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("repository: не удалось найти тестовый файл: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("repository: ожидается обычный файл, а не каталог или устройство")
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("repository: не удалось открыть тестовый файл: %w", err)
	}
	return file.Close()
}
