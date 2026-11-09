// Package app содержит CLI первых трёх этапов практической работы №2.
package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"dependencyvisualizer/src/config"
)

const (
	// ExitOK — код успешного завершения, в том числе при запросе справки.
	ExitOK = 0
	// ExitError — код ошибки аргументов или конфигурации.
	ExitError = 1
)

// Run загружает конфигурацию и запускает построение графа на этапе 3.
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
	if err := analyze(cfg, path, out); err != nil {
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
