package app

import (
	"github.com/xgamtx/go-mockery-descriptor/pkg/generator"
	"github.com/xgamtx/go-mockery-descriptor/pkg/parser"
)

func Run(dir, interfaceName string) (string, error) {
	desc, err := parser.ParseInterfaceInDir(dir, interfaceName)
	if err != nil {
		return "", err
	}

	return generator.Generate(desc), nil
}
