package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/RevoTale/no-js/internal/bundler/approutegen"
	"github.com/RevoTale/no-js/internal/bundler/templcssgen"
	bundlertemplgen "github.com/RevoTale/no-js/internal/bundler/templgen"
	"github.com/RevoTale/no-js/internal/projectlayout"
)

func main() {
	var rootDir string

	flag.StringVar(&rootDir, "root", ".", "application root directory")
	flag.Parse()

	layout, err := projectlayout.ResolveProjectLayoutFromRoot(rootDir)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "approutegen: %v\n", err)
		os.Exit(1)
	}

	if err := approutegen.Run(approutegen.Config{Layout: layout}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "approutegen: %v\n", err)
		os.Exit(1)
	}
	if err := templcssgen.Run(templcssgen.Config{Layout: layout}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "approutegen: generate templ css registry: %v\n", err)
		os.Exit(1)
	}
	if err := bundlertemplgen.Run(bundlertemplgen.Config{
		Paths:    []string{layout.GeneratedDir},
		BasePath: layout.RootDir,
	}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "approutegen: generate templ for generated routes: %v\n", err)
		os.Exit(1)
	}
}
