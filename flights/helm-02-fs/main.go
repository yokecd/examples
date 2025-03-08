package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/yokecd/yoke/pkg/flight"
	"github.com/yokecd/yoke/pkg/helm"
)

//go:embed all:redis
var fs embed.FS

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	chart, err := helm.LoadChartFromFS(fs)
	if err != nil {
		return fmt.Errorf("failed to load chart: %w", err)
	}

	resources, err := chart.Render(flight.Release(), flight.Namespace(), map[string]any{})
	if err != nil {
		return fmt.Errorf("failed to render chart: %w", err)
	}

	return json.NewEncoder(os.Stdout).Encode(resources)
}
