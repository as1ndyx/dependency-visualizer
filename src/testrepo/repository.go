// Package testrepo читает текстовый репозиторий для третьего этапа.
package testrepo

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

// MaxFileBytes ограничивает размер описания тестового репозитория.
const MaxFileBytes = 2 * 1024 * 1024

// Repository — пакеты с именами из заглавных латинских букв и их прямые зависимости.
type Repository map[string][]string

// Load открывает файл репозитория и проверяет его содержимое.
func Load(path string) (Repository, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("тестовый репозиторий: %w", err)
	}
	defer file.Close()
	return Parse(file)
}

// Parse читает строки вида A: B C; каждый пакет, включая листья, должен быть объявлен.
func Parse(input io.Reader) (Repository, error) {
	data, err := io.ReadAll(io.LimitReader(input, MaxFileBytes+1))
	if err != nil {
		return nil, fmt.Errorf("чтение тестового репозитория: %w", err)
	}
	if len(data) > MaxFileBytes {
		return nil, fmt.Errorf("тестовый репозиторий превышает %d байт", MaxFileBytes)
	}
	repository, err := parseLines(strings.TrimPrefix(string(data), "\uFEFF"))
	if err != nil {
		return nil, err
	}
	if err := repository.validateReferences(); err != nil {
		return nil, err
	}
	return repository, nil
}

// Dependencies возвращает копию списка зависимостей или ошибку неизвестного пакета.
func (r Repository) Dependencies(ctx context.Context, name string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	dependencies, exists := r[name]
	if !exists {
		return nil, fmt.Errorf("пакет %s не найден в тестовом репозитории", name)
	}
	return append([]string(nil), dependencies...), nil
}
