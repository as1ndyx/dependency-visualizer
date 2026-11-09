// Package config читает и проверяет INI-конфигурацию инструмента зависимостей.
package config

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

// Config — параметры приложения из секции [settings].
type Config struct {
	// Package — координаты Maven groupId:artifactId:version или имя тестового пакета.
	Package string
	// Repository — URL репозитория или путь к тестовому файлу, как записано в INI.
	Repository string
	// TestMode переключает проверку URL на проверку локального файла.
	TestMode bool
	// Filter — подстрока для исключения пакетов при обходе графа на этапе 3.
	// Пустая строка означает, что фильтрация выключена.
	Filter string
}

// Pair — один параметр конфигурации в формате ключ-значение.
type Pair struct {
	Key   string
	Value string
}

// Load читает INI-файл и проверяет все четыре обязательных параметра.
// Относительный путь к тестовому репозиторию считается от каталога INI-файла.
func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("config: не удалось открыть файл: %w", err)
	}
	defer file.Close()
	values, err := parseINI(file)
	if err != nil {
		return Config{}, err
	}
	return fromValues(values, path)
}

// fromValues проверяет обязательные ключи, тип режима и значения параметров.
func fromValues(values map[string]string, path string) (Config, error) {
	for _, key := range []string{"package", "repository", "test_mode", "filter"} {
		if _, ok := values[key]; !ok {
			return Config{}, fmt.Errorf("%s: обязательный параметр отсутствует", key)
		}
	}
	mode := values["test_mode"]
	if mode != "true" && mode != "false" {
		return Config{}, fmt.Errorf("test_mode: ожидается true или false")
	}
	cfg := Config{values["package"], values["repository"], mode == "true", values["filter"]}
	if err := cfg.validate(path); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Pairs возвращает параметры в постоянном порядке для отладочного вывода.
func (c Config) Pairs() []Pair {
	return []Pair{
		{Key: "package", Value: c.Package},
		{Key: "repository", Value: c.Repository},
		{Key: "test_mode", Value: strconv.FormatBool(c.TestMode)},
		{Key: "filter", Value: c.Filter},
	}
}

// PrintDebug выводит все параметры при запуске, как требует первый этап.
func PrintDebug(out io.Writer, cfg Config) error {
	for _, pair := range cfg.Pairs() {
		if _, err := fmt.Fprintf(out, "%s = %s\n", pair.Key, pair.Value); err != nil {
			return err
		}
	}
	return nil
}
