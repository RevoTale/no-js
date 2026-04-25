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

	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "usage: templcssgen [-root .] [-config path]\n\n")
		_, _ = fmt.Fprintln(
			flag.CommandLine.Output(),
			"internal helper; app projects should use `go tool no-js gen routes -root .`",
		)
		_, _ = fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}
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
