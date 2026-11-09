// Package app содержит CLI первых четырёх этапов практической работы №2.
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

// options задаёт путь к INI и необязательный режим обратных зависимостей этапа 4.
type options struct {
	ConfigPath string
	Reverse    string
}

// Run загружает конфигурацию и запускает обычный или обратный анализ графа.
func Run(args []string, out, errOut io.Writer) int {
	options, err := parseFlags(args, errOut)
	if errors.Is(err, flag.ErrHelp) {
		return ExitOK
	}
	if err != nil {
		fmt.Fprintln(errOut, "ошибка:", err)
		return ExitError
	}
	cfg, err := config.Load(options.ConfigPath)
	if err != nil {
		fmt.Fprintln(errOut, "ошибка:", err)
		return ExitError
	}
	if err := analyze(cfg, options, out); err != nil {
		fmt.Fprintln(errOut, "ошибка:", err)
		return ExitError
	}
	return ExitOK
}

// parseFlags принимает путь к INI и выбранный для обратного анализа пакет.
func parseFlags(args []string, out io.Writer) (options, error) {
	set := flag.NewFlagSet("depgraph", flag.ContinueOnError)
	set.SetOutput(out)
	path := set.String("config", "config.ini", "путь к конфигурационному файлу INI")
	reverse := set.String("reverse", "", "вывести обратные зависимости указанного пакета")
	if err := set.Parse(args); err != nil {
		return options{}, err
	}
	if set.NArg() != 0 {
		return options{}, fmt.Errorf("лишние аргументы: %s", strings.Join(set.Args(), " "))
	}
	if strings.TrimSpace(*path) == "" {
		return options{}, fmt.Errorf("config: путь к конфигурации не должен быть пустым")
	}
	reverseSet := false
	set.Visit(func(item *flag.Flag) {
		if item.Name == "reverse" {
			reverseSet = true
		}
	})
	if reverseSet && strings.TrimSpace(*reverse) == "" {
		return options{}, fmt.Errorf("reverse: имя пакета не должно быть пустым")
	}
	return options{ConfigPath: *path, Reverse: *reverse}, nil
}
