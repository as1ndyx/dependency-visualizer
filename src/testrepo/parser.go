package testrepo

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var packageName = regexp.MustCompile(`^[A-Z]+$`)

// parseLines пропускает пустые строки и полнострочные комментарии # и ;.
func parseLines(text string) (Repository, error) {
	repository := make(Repository)
	for number, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if err := parseLine(line, repository); err != nil {
			return nil, fmt.Errorf("тестовый репозиторий, строка %d: %w", number+1, err)
		}
	}
	if len(repository) == 0 {
		return nil, fmt.Errorf("тестовый репозиторий не содержит пакетов")
	}
	return repository, nil
}

// parseLine проверяет имена, запрещает повторные объявления и убирает повторные рёбра.
func parseLine(line string, repository Repository) error {
	name, rest, ok := strings.Cut(line, ":")
	name = strings.TrimSpace(name)
	if !ok || !packageName.MatchString(name) {
		return fmt.Errorf("ожидается запись A: B C с заглавными латинскими именами")
	}
	if _, exists := repository[name]; exists {
		return fmt.Errorf("пакет %s объявлен повторно", name)
	}
	var dependencies []string
	seen := make(map[string]bool)
	for _, dependency := range strings.Fields(rest) {
		if !packageName.MatchString(dependency) {
			return fmt.Errorf("неверное имя зависимости %q", dependency)
		}
		if !seen[dependency] {
			dependencies = append(dependencies, dependency)
			seen[dependency] = true
		}
	}
	repository[name] = dependencies
	return nil
}

// validateReferences запрещает ссылки на необъявленные пакеты, включая недостижимую часть файла.
// Сортировка обеспечивает одинаковую первую ошибку при нескольких неверных ссылках.
func (r Repository) validateReferences() error {
	names := make([]string, 0, len(r))
	for name := range r {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		for _, dependency := range r[name] {
			if _, exists := r[dependency]; !exists {
				return fmt.Errorf("тестовый репозиторий: %s ссылается на необъявленный пакет %s", name, dependency)
			}
		}
	}
	return nil
}
