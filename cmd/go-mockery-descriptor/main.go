package main

import (
	"bytes"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/xgamtx/go-mockery-descriptor/internal/app"
	"github.com/xgamtx/go-mockery-descriptor/internal/config"
)

func initConfig() *config.Config {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	return cfg
}

func generateFileName(cfg *config.Config, interfaceName string) (string, error) {
	tmpl := template.New("fileName.tmpl")
	tmpl, err := tmpl.Parse(cfg.Output)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, strings.ToLower(interfaceName)); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func main() {
	cfg := initConfig()
	for _, ifaceCfg := range cfg.Interfaces {
		output, err := app.Run(&ifaceCfg)
		if err != nil {
			log.Fatalf("Failed to generate code: %v", err)
		}

		fileName, err := generateFileName(cfg, ifaceCfg.Name)
		if err != nil {
			log.Fatalf("Failed to generate code: %v", err)
		}

		if err = os.WriteFile(fileName, []byte(output), 0o600); err != nil { //nolint:mnd
			log.Fatalf("Failed to write output file: %v", err)
		}
	}
}
