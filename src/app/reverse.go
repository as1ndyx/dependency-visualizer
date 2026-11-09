package app

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"dependencyvisualizer/src/config"
	"dependencyvisualizer/src/graph"
	"dependencyvisualizer/src/testrepo"
)

// analyzeReverse находит обратные зависимости с помощью DFS предыдущего этапа.
func analyzeReverse(ctx context.Context, cfg config.Config, options options, out io.Writer) error {
	if cfg.Filter != "" && strings.Contains(options.Reverse, cfg.Filter) {
		return fmt.Errorf("reverse: пакет %s исключён фильтром", options.Reverse)
	}
	edges, scope, err := reverseArea(ctx, cfg, options.ConfigPath)
	if err != nil {
		return err
	}
	result, err := graph.Reverse(ctx, edges, options.Reverse, cfg.Filter)
	if err != nil {
		return err
	}
	if err := graph.PrintReverse(out, result, scope); err != nil {
		return fmt.Errorf("ошибка вывода: %w", err)
	}
	return nil
}

// reverseArea использует весь тестовый файл либо доступную из package часть Maven-графа.
// Один POM не содержит списка всех пакетов репозитория, зависящих от него.
func reverseArea(ctx context.Context, cfg config.Config, configPath string) (map[string][]string, string, error) {
	if cfg.TestMode {
		repository, err := loadTestRepository(cfg, configPath)
		return repository, "весь тестовый репозиторий", err
	}
	source, err := dependencySource(cfg, configPath)
	if err != nil {
		return nil, "", err
	}
	result, err := graph.Build(ctx, cfg.Package, cfg.Filter, source)
	return result.Edges, "граф " + cfg.Package, err
}

// loadTestRepository единообразно разрешает путь относительно INI для обоих режимов вывода.
func loadTestRepository(cfg config.Config, configPath string) (testrepo.Repository, error) {
	path := cfg.Repository
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(configPath), path)
	}
	return testrepo.Load(path)
}
