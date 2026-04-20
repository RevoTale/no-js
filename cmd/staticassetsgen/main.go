package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/RevoTale/no-js/internal/bundler/staticassets"
	"github.com/RevoTale/no-js/internal/bundler/templcssgen"
	"github.com/RevoTale/no-js/internal/filesystem"
	"github.com/RevoTale/no-js/internal/projectlayout"
)

func main() {
	var sourceDir string
	var outDir string
	var manifestPath string
	var urlPrefix string
	var rootDir string
	var configPath string
	var templCSS bool

	flag.StringVar(&sourceDir, "source", "web/assets", "source static directory")
	flag.StringVar(&outDir, "out", "web/assets-build", "output static directory")
	flag.StringVar(&manifestPath, "manifest", "web/assets-build/manifest.json", "manifest output path")
	flag.StringVar(&urlPrefix, "url-prefix", "/_assets/", "base static URL prefix")
	flag.StringVar(&rootDir, "root", ".", "application root directory")
	flag.StringVar(&configPath, "config", "", "bundle config path")
	flag.BoolVar(&templCSS, "templ-css", false, "generate styles/templ.css before bundling")
	flag.Parse()

	var layout projectlayout.ProjectLayout
	if templCSS {
		var err error
		layout, err = resolveLayout(rootDir, configPath)
		if err != nil {
			exitf("%v", err)
		}
		if sourceDir == "web/assets" {
			sourceDir = layout.StaticAssets.SourceDir
		}
		if outDir == "web/assets-build" {
			outDir = layout.StaticAssets.OutDir
		}
		if manifestPath == "web/assets-build/manifest.json" {
			manifestPath = layout.StaticAssets.ManifestPath
		}
	}
	sourceDir = resolvePath(rootDir, sourceDir)
	outDir = resolvePath(rootDir, outDir)
	manifestPath = resolvePath(rootDir, manifestPath)

	buildSourceDir := sourceDir
	cleanupSource := func() error { return nil }
	if templCSS {
		stageDir, cleanup, err := templcssgen.PrepareStaticSource(templcssgen.PrepareStaticSourceConfig{
			Layout:    layout,
			SourceDir: sourceDir,
		})
		if err != nil {
			exitf("prepare templ css static source: %v", err)
		}
		buildSourceDir = stageDir
		cleanupSource = cleanup
	}
	defer func() {
		if cleanupErr := cleanupSource(); cleanupErr != nil {
			exitf("cleanup staged source: %v", cleanupErr)
		}
	}()

	bundle, err := staticassets.Build(staticassets.BuildConfig{
		SourceDir: buildSourceDir,
		URLPrefix: urlPrefix,
	})
	if err != nil {
		exitf("build static bundle: %v", err)
	}
	defer func() {
		if cleanupErr := bundle.Cleanup(); cleanupErr != nil {
			exitf("cleanup temp bundle: %v", cleanupErr)
		}
	}()

	if err := os.RemoveAll(outDir); err != nil {
		exitf("clean output dir %q: %v", outDir, err)
	}
	if err := filesystem.CopyTree(bundle.Dir(), outDir); err != nil {
		exitf("copy processed assets to %q: %v", outDir, err)
	}

	if strings.TrimSpace(manifestPath) == "" {
		manifestPath = filepath.Join(outDir, "manifest.json")
	}
	if err := staticassets.WriteManifest(manifestPath, bundle.Manifest()); err != nil {
		exitf("write manifest %q: %v", manifestPath, err)
	}
}

func resolveLayout(rootDir string, configPath string) (projectlayout.ProjectLayout, error) {
	if strings.TrimSpace(configPath) == "" {
		return projectlayout.ResolveProjectLayoutFromRoot(rootDir)
	}

	cfg, err := projectlayout.LoadConfigFile(configPath)
	if err != nil {
		return projectlayout.ProjectLayout{}, err
	}
	return projectlayout.ResolveProjectLayout(rootDir, cfg)
}

func resolvePath(rootDir string, target string) string {
	if filepath.IsAbs(target) {
		return target
	}
	return filepath.Join(rootDir, target)
}

func exitf(formatText string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, formatText+"\n", args...)
	os.Exit(1)
}
