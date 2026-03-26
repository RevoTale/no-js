package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/RevoTale/no-js/bundler"
	"github.com/RevoTale/no-js/bundler/approutegen"
)

func main() {
	var rootDir string

	flag.StringVar(&rootDir, "root", ".", "application root directory")
	flag.Parse()

	layout, err := bundler.ResolveProjectLayoutFromRoot(rootDir)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "approutegen: %v\n", err)
		os.Exit(1)
	}

	if err := approutegen.Run(approutegen.Config{Layout: layout}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "approutegen: %v\n", err)
		os.Exit(1)
	}
}
