package app

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"dependencyvisualizer/src/config"
	"dependencyvisualizer/src/graph"
	"dependencyvisualizer/src/maven"
	"dependencyvisualizer/src/testrepo"
)

// analyze строит и выводит граф, ограничивая общее время анализа.
func analyze(cfg config.Config, configPath string, out io.Writer) error {
	const analysisTimeout = 60 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), analysisTimeout)
	defer cancel()
	source, err := dependencySource(cfg, configPath)
	if err != nil {
		return err
	}
	result, err := graph.Build(ctx, cfg.Package, cfg.Filter, source)
	if err != nil {
		return err
	}
	if err := graph.Print(out, result); err != nil {
		return fmt.Errorf("ошибка вывода: %w", err)
	}
	return nil
}

// dependencySource выбирает источник; относительный путь считается от INI, а не от cwd.
func dependencySource(cfg config.Config, configPath string) (graph.Source, error) {
	if cfg.TestMode {
		path := cfg.Repository
		if !filepath.IsAbs(path) {
			path = filepath.Join(filepath.Dir(configPath), path)
		}
		repository, err := testrepo.Load(path)
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
