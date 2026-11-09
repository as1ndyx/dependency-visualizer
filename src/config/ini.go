package config

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// parseINI читает секцию [settings], пропуская пустые строки и комментарии.
// На этапе 1 используется небольшой INI-парсер без сторонних библиотек.
func parseINI(in io.Reader) (map[string]string, error) {
	values := make(map[string]string)
	scanner := bufio.NewScanner(in)
	section := false
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\uFEFF"))
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if err := parseLine(line, &section, values); err != nil {
			return nil, fmt.Errorf("INI, строка %d: %w", lineNumber, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("config: ошибка чтения INI: %w", err)
	}
	if !section {
		return nil, fmt.Errorf("INI: отсутствует секция [settings]")
	}
	return values, nil
}

// parseLine проверяет заголовок секции либо сохраняет пару ключ-значение.
func parseLine(line string, section *bool, values map[string]string) error {
	if strings.HasPrefix(line, "[") {
		if line != "[settings]" {
			return fmt.Errorf("неизвестная секция %q, ожидается [settings]", line)
		}
		if *section {
			return fmt.Errorf("секция [settings] указана повторно")
		}
		*section = true
		return nil
	}
	if !*section {
		return fmt.Errorf("параметр задан вне секции [settings]")
	}
	key, value, ok := strings.Cut(line, "=")
	if !ok {
		return fmt.Errorf("ожидается запись ключ = значение")
	}
	return storeValue(strings.TrimSpace(key), strings.TrimSpace(value), values)
}

// storeValue отклоняет опечатки, повторы ключей и непечатные значения.
// Значения не заключаются в кавычки; символы #, ; и = внутри них сохраняются.
func storeValue(key, value string, values map[string]string) error {
	switch key {
	case "package", "repository", "test_mode", "filter":
	default:
		return fmt.Errorf("неизвестный параметр %q", key)
	}
	if _, exists := values[key]; exists {
		return fmt.Errorf("%s: параметр указан повторно", key)
	}
	if !utf8.ValidString(value) || strings.ContainsFunc(value, unicode.IsControl) {
		return fmt.Errorf("%s: ожидается текст UTF-8 без управляющих символов", key)
	}
	values[key] = value
	return nil
}
