package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/RevoTale/no-js/internal/bundler/i18nkeygen"
)

func main() {
	var inputPath string
	var outputPath string
	var packageName string

	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "usage: i18nkeygen -in messages.json -out keys_gen.go [-pkg i18n]\n\n")
		_, _ = fmt.Fprintln(
			flag.CommandLine.Output(),
			"internal helper; app projects should use `go tool no-js gen routes -root .`",
		)
		_, _ = fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}
	flag.StringVar(&inputPath, "in", "", "canonical locale json file path")
	flag.StringVar(&outputPath, "out", "", "generated go file output path")
	flag.StringVar(&packageName, "pkg", "i18n", "generated go package name")
	flag.Parse()

	inputPath = strings.TrimSpace(inputPath)
	if inputPath == "" {
		exitf("missing -in")
	}
	outputPath = strings.TrimSpace(outputPath)
	if outputPath == "" {
		exitf("missing -out")
	}

	source, err := os.ReadFile(inputPath)
	if err != nil {
		exitf("read %q: %v", inputPath, err)
	}

	generatedSource, err := i18nkeygen.GenerateFromJSON(packageName, source)
	if err != nil {
		exitf("generate keys: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		exitf("create output directory for %q: %v", outputPath, err)
	}
	if err := os.WriteFile(outputPath, generatedSource, 0o644); err != nil {
		exitf("write %q: %v", outputPath, err)
	}
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
