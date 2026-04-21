package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/RevoTale/no-js/internal/bundler/templcssgen"
	"github.com/RevoTale/no-js/internal/projectlayout"
)

func main() {
	var rootDir string
	var configPath string

	flag.StringVar(&rootDir, "root", ".", "application root directory")
	flag.StringVar(&configPath, "config", "", "bundle config path")
	flag.Parse()

	layout, err := resolveLayout(rootDir, configPath)
	if err != nil {
		exitf("%v", err)
	}
	if err := templcssgen.Run(templcssgen.Config{Layout: layout}); err != nil {
		exitf("generate templ css registry: %v", err)
	}
}

func resolveLayout(rootDir string, configPath string) (projectlayout.ProjectLayout, error) {
	if configPath == "" {
		return projectlayout.ResolveProjectLayoutFromRoot(rootDir)
	}

	cfg, err := projectlayout.LoadConfigFile(configPath)
	if err != nil {
		return projectlayout.ProjectLayout{}, err
	}
	return projectlayout.ResolveProjectLayout(rootDir, cfg)
}

func exitf(formatText string, args ...any) {
	fmt.Fprintf(os.Stderr, formatText+"\n", args...)
	os.Exit(1)
}
