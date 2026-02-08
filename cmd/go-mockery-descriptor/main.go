package main

import (
	"log"
	"os"
	"strings"

	"github.com/xgamtx/go-mockery-descriptor/internal/app"
	"github.com/xgamtx/go-mockery-descriptor/internal/config"
)

func initConfig() *config.Config {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	if cfg.Output == "" {
		cfg.Output = strings.ToLower(cfg.Interface) + ".gen.go"
	}

	if cfg.Dir == "" || cfg.Interface == "" {
		log.Fatalf("Usage: %s --dir=<path_to_file> --interface=<interface_name>\n", os.Args[0])
	}

	return cfg
}

func main() {
	cfg := initConfig()
	output, err := app.Run(cfg)
	if err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}

	if err = os.WriteFile(cfg.Output, []byte(output), 0o600); err != nil { //nolint:mnd
		log.Fatalf("Failed to write output file: %v", err)
	}
}
