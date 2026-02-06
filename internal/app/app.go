package app

import (
	"github.com/xgamtx/go-mockery-descriptor/internal/fieldoverwriter"
	"github.com/xgamtx/go-mockery-descriptor/internal/generator"
	"github.com/xgamtx/go-mockery-descriptor/internal/parser"
)

func Run(dir, interfaceName string, fieldOverwriterParams []string) (string, error) {
	desc, err := parser.ParseInterfaceInDir(dir, interfaceName)
	if err != nil {
		return "", err
	}

	overwriterStorage, err := fieldoverwriter.NewStorage(fieldOverwriterParams)
	if err != nil {
		return "", err
	}

	return generator.Generate(desc, overwriterStorage)
}
