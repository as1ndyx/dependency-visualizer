package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"dependencyvisualizer/src/config"
	"dependencyvisualizer/src/graph"
	"dependencyvisualizer/src/maven"
)

// analyze запускает выбранную операцию, ограничивая общее время анализа.
func analyze(cfg config.Config, options options, out io.Writer) error {
	const analysisTimeout = 60 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), analysisTimeout)
	defer cancel()
	if options.Reverse != "" {
		return analyzeReverse(ctx, cfg, options, out)
	}
	source, err := dependencySource(cfg, options.ConfigPath)
	if err != nil {
		return err
	}
	result, err := graph.Build(ctx, cfg.Package, cfg.Filter, source)
	if err != nil {
		return err
	}
	printGraph := graph.PrintDOT
	if options.Format == "text" {
		printGraph = graph.Print
	}
	if err := printGraph(out, result); err != nil {
		return fmt.Errorf("ошибка вывода: %w", err)
	}
	return nil
}

// dependencySource выбирает источник; относительный путь считается от INI, а не от cwd.
func dependencySource(cfg config.Config, configPath string) (graph.Source, error) {
	if cfg.TestMode {
		repository, err := loadTestRepository(cfg, configPath)
		if err != nil {
			return nil, err
		}
		return repository.Dependencies, nil
	}
	client := maven.NewClient(nil)
	return func(ctx context.Context, name string) ([]string, error) {
		dependencies, err := client.DirectDependencies(ctx, cfg.Repository, name)
		if err != nil {
			return nil, err
		}
		names := make([]string, 0, len(dependencies))
		for _, dependency := range dependencies {
			names = append(names, dependency.Coordinates.String())
		}
		return names, nil
	}, nil
}
