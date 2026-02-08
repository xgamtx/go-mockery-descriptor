package app

import (
	"github.com/xgamtx/go-mockery-descriptor/internal/config"
	"github.com/xgamtx/go-mockery-descriptor/internal/fieldoverwriter"
	"github.com/xgamtx/go-mockery-descriptor/internal/generator"
	"github.com/xgamtx/go-mockery-descriptor/internal/parser"
)

func Run(cfg *config.Config) (string, error) {
	desc, err := parser.ParseInterfaceInDir(cfg.Dir, cfg.Interface)
	if err != nil {
		return "", err
	}

	overwriterStorage, err := fieldoverwriter.NewStorage(cfg.FieldOverwriterParams)
	if err != nil {
		return "", err
	}

	return generator.Generate(desc, overwriterStorage)
}
