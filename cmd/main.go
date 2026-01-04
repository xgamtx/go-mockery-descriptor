package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/xgamtx/go-mockery-descriptor/pkg/generator"
	"github.com/xgamtx/go-mockery-descriptor/pkg/parser"
)

type args struct {
	fileName       string
	interfaceName  string
	outputFileName string
}

func getArgs() args {
	var args args
	flag.StringVar(&args.fileName, "file-name", "", "filename to look for an interface")
	flag.StringVar(&args.interfaceName, "interface", "", "interface name")
	flag.StringVar(&args.outputFileName, "output", "", "target file name")
	flag.Parse()

	if args.outputFileName == "" {
		args.outputFileName = strings.ToLower(args.interfaceName) + ".gen.go"
	}

	if args.fileName == "" || args.interfaceName == "" {
		log.Fatalf("Usage: %s --file-name=<path_to_file> --interface=<interface_name> \n", os.Args[0])
	}

	return args
}

func main() {
	args := getArgs()
	desc, err := parser.ParseInterfaceInDir(args.fileName, args.interfaceName)
	if err != nil {
		log.Fatalf("Failed to parse interface: %v", err)
	}

	output := generator.Generate(desc)
	if err = os.WriteFile(args.outputFileName, []byte(output), 0o600); err != nil { //nolint:mnd
		log.Fatalf("Failed to write output file: %v", err)
	}
}
