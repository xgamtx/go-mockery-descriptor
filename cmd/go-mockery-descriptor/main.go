package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/xgamtx/go-mockery-descriptor/internal/app"
)

type StringSlice []string

func (ss *StringSlice) String() string {
	return fmt.Sprintf("%v", *ss)
}

func (ss *StringSlice) Set(value string) error {
	*ss = append(*ss, value)

	return nil
}

type args struct {
	dir                   string
	interfaceName         string
	outputFileName        string
	fieldOverwriterParams StringSlice
}

func getArgs() args {
	var args args
	flag.StringVar(&args.dir, "dir", ".", "directory to parse")
	flag.StringVar(&args.interfaceName, "interface", "", "interface name")
	flag.StringVar(&args.outputFileName, "output", "", "target file name")
	flag.Var(&args.fieldOverwriterParams, "field-overwriter-param", "field overwriter param, can be used more than once")
	flag.Parse()

	if args.outputFileName == "" {
		args.outputFileName = strings.ToLower(args.interfaceName) + ".gen.go"
	}

	if args.dir == "" || args.interfaceName == "" {
		log.Fatalf("Usage: %s --dir=<path_to_file> --interface=<interface_name>\n", os.Args[0])
	}

	return args
}

func main() {
	args := getArgs()
	output, err := app.Run(args.dir, args.interfaceName, args.fieldOverwriterParams)
	if err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}

	if err = os.WriteFile(args.outputFileName, []byte(output), 0o600); err != nil { //nolint:mnd
		log.Fatalf("Failed to write output file: %v", err)
	}
}
