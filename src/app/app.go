// Package app содержит CLI всех пяти этапов практической работы №2.
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

// options задаёт путь к INI, режим этапа 4 и формат вывода этапа 5.
type options struct {
	ConfigPath string
	Reverse    string
	Format     string
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

// parseFlags принимает путь к INI, формат вывода и необязательный обратный анализ.
func parseFlags(args []string, out io.Writer) (options, error) {
	var result options
	set := flag.NewFlagSet("depgraph", flag.ContinueOnError)
	set.SetOutput(out)
	set.StringVar(&result.ConfigPath, "config", "config.ini", "путь к конфигурационному файлу INI")
	set.StringVar(&result.Reverse, "reverse", "", "вывести обратные зависимости указанного пакета")
	set.StringVar(&result.Format, "format", "dot", "формат вывода графа: dot или text")
	if err := set.Parse(args); err != nil {
		return options{}, err
	}
	if set.NArg() != 0 {
		return options{}, fmt.Errorf("лишние аргументы: %s", strings.Join(set.Args(), " "))
	}
	explicit := make(map[string]bool)
	set.Visit(func(item *flag.Flag) { explicit[item.Name] = true })
	return validateOptions(result, explicit)
}

// validateOptions проверяет формат; режим -reverse по умолчанию сохраняет текстовый вывод этапа 4.
func validateOptions(result options, explicit map[string]bool) (options, error) {
	if strings.TrimSpace(result.ConfigPath) == "" {
		return options{}, fmt.Errorf("config: путь к конфигурации не должен быть пустым")
	}
	if explicit["reverse"] && strings.TrimSpace(result.Reverse) == "" {
		return options{}, fmt.Errorf("reverse: имя пакета не должно быть пустым")
	}
	if result.Reverse != "" && !explicit["format"] {
		result.Format = "text"
	}
	if result.Format != "dot" && result.Format != "text" {
		return options{}, fmt.Errorf("format: ожидается dot или text")
	}
	if result.Reverse != "" && result.Format != "text" {
		return options{}, fmt.Errorf("reverse: для обратного анализа используется -format text")
	}
	return result, nil
}
