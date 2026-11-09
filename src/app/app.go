// Package app содержит CLI первых двух этапов практической работы №2.
package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"dependencyvisualizer/src/config"
	"dependencyvisualizer/src/maven"
)

const (
	// ExitOK — код успешного завершения, в том числе при запросе справки.
	ExitOK = 0
	// ExitError — код ошибки аргументов или конфигурации.
	ExitError = 1
)

// Run загружает конфигурацию и запускает сбор прямых зависимостей на этапе 2.
func Run(args []string, out, errOut io.Writer) int {
	path, err := parseFlags(args, errOut)
	if errors.Is(err, flag.ErrHelp) {
		return ExitOK
	}
	if err != nil {
		fmt.Fprintln(errOut, "ошибка:", err)
		return ExitError
	}
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintln(errOut, "ошибка:", err)
		return ExitError
	}
	if err := analyze(cfg, out); err != nil {
		fmt.Fprintln(errOut, "ошибка:", err)
		return ExitError
	}
	return ExitOK
}

// parseFlags принимает только путь к INI-файлу: сами настройки читаются из него.
func parseFlags(args []string, out io.Writer) (string, error) {
	set := flag.NewFlagSet("depgraph", flag.ContinueOnError)
	set.SetOutput(out)
	path := set.String("config", "config.ini", "путь к конфигурационному файлу INI")
	if err := set.Parse(args); err != nil {
		return "", err
	}
	if set.NArg() != 0 {
		return "", fmt.Errorf("лишние аргументы: %s", strings.Join(set.Args(), " "))
	}
	if strings.TrimSpace(*path) == "" {
		return "", fmt.Errorf("config: путь к конфигурации не должен быть пустым")
	}
	return *path, nil
}

// analyze ограничивает общее время анализа, включая родителей и BOM.
// Фильтрация и разбор тестового графа будут добавлены на этапе 3.
func analyze(cfg config.Config, out io.Writer) error {
	if cfg.TestMode {
		return fmt.Errorf("test_mode: анализ тестового графа будет доступен на этапе 3; для Maven задайте false")
	}
	const analysisTimeout = 60 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), analysisTimeout)
	defer cancel()
	dependencies, err := maven.NewClient(nil).DirectDependencies(ctx, cfg.Repository, cfg.Package)
	if err != nil {
		return err
	}
	if err := maven.PrintDependencies(out, cfg.Package, dependencies); err != nil {
		return fmt.Errorf("ошибка вывода: %w", err)
	}
	return nil
}
