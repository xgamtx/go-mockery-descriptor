package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/xgamtx/go-mockery-descriptor/pkg/app"
)

type args struct {
	dir                   string
	interfaceName         string
	outputFileName        string
	fieldOverwriterParams []string
	fullPackagePath       string
}

func getArgs() args {
	var args args
	flag.StringVar(&args.dir, "dir", ".", "directory to parse")
	flag.StringVar(&args.interfaceName, "interface", "", "interface name")
	flag.StringVar(&args.outputFileName, "output", "", "target file name")
	flag.StringVar(&args.fullPackagePath, "full-package-path", "", "package name")
	var fieldOverwriterParam string
	flag.StringVar(&fieldOverwriterParam, "field-overwriter-param", "", "field overwriter param")
	flag.Parse()

	if fieldOverwriterParam != "" {
		args.fieldOverwriterParams = []string{fieldOverwriterParam}
	}

	if args.outputFileName == "" {
		args.outputFileName = strings.ToLower(args.interfaceName) + ".gen.go"
	}

	if args.dir == "" || args.interfaceName == "" || args.fullPackagePath == "" {
		log.Fatalf("Usage: %s --dir=<path_to_file> --interface=<interface_name> --full-package-path=<full_package_path>\n", os.Args[0])
	}

	return args
}

func main() {
	args := getArgs()
	output, err := app.Run(args.dir, args.interfaceName, args.fieldOverwriterParams, args.fullPackagePath)
	if err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}

	if err = os.WriteFile(args.outputFileName, []byte(output), 0o600); err != nil { //nolint:mnd
		log.Fatalf("Failed to write output file: %v", err)
	}
}
