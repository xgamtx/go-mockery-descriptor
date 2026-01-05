package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/xgamtx/go-mockery-descriptor/pkg/app"
)

type args struct {
	dir            string
	interfaceName  string
	outputFileName string
}

func getArgs() args {
	var args args
	flag.StringVar(&args.dir, "dir", ".", "directory to parse")
	flag.StringVar(&args.interfaceName, "interface", "", "interface name")
	flag.StringVar(&args.outputFileName, "output", "", "target file name")
	flag.Parse()

	if args.outputFileName == "" {
		args.outputFileName = strings.ToLower(args.interfaceName) + ".gen.go"
	}

	if args.dir == "" || args.interfaceName == "" {
		log.Fatalf("Usage: %s --file-name=<path_to_file> --interface=<interface_name> \n", os.Args[0])
	}

	return args
}

func main() {
	args := getArgs()
	output, err := app.Run(args.dir, args.interfaceName)
	if err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}

	if err = os.WriteFile(args.outputFileName, []byte(output), 0o600); err != nil { //nolint:mnd
		log.Fatalf("Failed to write output file: %v", err)
	}
}
