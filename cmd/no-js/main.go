package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/RevoTale/no-js/internal/bundler/approutegen"
	"github.com/RevoTale/no-js/internal/bundler/clientassets"
	"github.com/RevoTale/no-js/internal/bundler/i18ngen"
	bundlerstaticassets "github.com/RevoTale/no-js/internal/bundler/staticassets"
	"github.com/RevoTale/no-js/internal/bundler/templcssgen"
	bundlertemplgen "github.com/RevoTale/no-js/internal/bundler/templgen"
	"github.com/RevoTale/no-js/internal/filesystem"
	"github.com/RevoTale/no-js/internal/projectlayout"
)

func main() {
	if len(os.Args) < 2 {
		exitf("usage: no-js gen [routes|assets|check] [-root .] [-config path]")
	}

	switch os.Args[1] {
	case "gen":
		if err := runGen(os.Args[2:]); err != nil {
			exitf("%v", err)
		}
	default:
		exitf("unknown command %q", os.Args[1])
	}
}

func runGen(args []string) error {
	mode := "all"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		mode = strings.TrimSpace(args[0])
		args = args[1:]
	}

	switch mode {
	case "", "all":
		mode = "all"
	case "routes", "assets", "check":
	default:
		return fmt.Errorf("unknown gen mode %q", mode)
	}

	flags := flag.NewFlagSet("gen", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	var rootDir string
	var configPath string
	flags.StringVar(&rootDir, "root", ".", "application root directory")
	flags.StringVar(&configPath, "config", "", "bundle config path")
	if err := flags.Parse(args); err != nil {
		return err
	}

	layout, err := resolveLayout(rootDir, configPath)
	if err != nil {
		return err
	}

	switch mode {
	case "routes":
		return generateRoutes(layout)
	case "assets":
		return generateAssets(layout)
	case "check":
		if err := generateRoutes(layout); err != nil {
			return err
		}
		if err := generateAssets(layout); err != nil {
			return err
		}
		return checkGitDiff(layout.RootDir)
	default:
		if err := generateRoutes(layout); err != nil {
			return err
		}
		return generateAssets(layout)
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
	layout, err := projectlayout.ResolveProjectLayout(rootDir, cfg)
	if err != nil {
		return projectlayout.ProjectLayout{}, err
	}
	resolvedConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return projectlayout.ProjectLayout{}, fmt.Errorf("resolve config path %q: %w", configPath, err)
	}
	layout.ConfigPath = resolvedConfigPath
	return layout, nil
}

func generateRoutes(layout projectlayout.ProjectLayout) error {
	if err := approutegen.Run(approutegen.Config{Layout: layout}); err != nil {
		return fmt.Errorf("generate routes: %w", err)
	}
	if err := i18ngen.Run(i18ngen.Config{Layout: layout}); err != nil {
		return fmt.Errorf("generate i18n: %w", err)
	}
	if layout.Assets.TemplCSS {
		if err := templcssgen.Run(templcssgen.Config{Layout: layout}); err != nil {
			return fmt.Errorf("generate templ css registry: %w", err)
		}
	} else if err := templcssgen.Cleanup(templcssgen.Config{Layout: layout}); err != nil {
		return fmt.Errorf("clean templ css registry: %w", err)
	}
	if err := generateTemplInDir(layout.RootDir, layout.GeneratedDir); err != nil {
		return fmt.Errorf("generate templ for generated routes: %w", err)
	}
	return nil
}

func generateAssets(layout projectlayout.ProjectLayout) error {
	templCSSHasRegistrations := false
	if layout.Assets.TemplCSS {
		var err error
		templCSSHasRegistrations, err = templcssgen.HasRegistrations(templcssgen.Config{Layout: layout})
		if err != nil {
			return fmt.Errorf("inspect templ css registrations: %w", err)
		}
	}

	hasClientAssets, err := clientassets.HasClientAssets(layout)
	if err != nil {
		return fmt.Errorf("discover client assets: %w", err)
	}
	if !layout.ServerFeatures.StaticAssets && !templCSSHasRegistrations && !hasClientAssets {
		return cleanStaticAssetOutput(layout)
	}

	sourceDir := strings.TrimSpace(layout.StaticAssets.SourceDir)
	if sourceDir == "" {
		return errors.New("static assets source dir is required")
	}
	outDir := strings.TrimSpace(layout.StaticAssets.OutDir)
	if outDir == "" {
		return errors.New("static assets out dir is required")
	}
	manifestPath := strings.TrimSpace(layout.StaticAssets.ManifestPath)
	if manifestPath == "" {
		manifestPath = filepath.Join(outDir, "manifest.json")
	}

	buildSourceDir := sourceDir
	cleanupSource := func() error { return nil }
	defer func() {
		_ = cleanupSource()
	}()

	templCSSStylesheetBuilt := false
	if templCSSHasRegistrations {
		stageDir, cleanup, err := templcssgen.PrepareStaticSource(templcssgen.PrepareStaticSourceConfig{
			Layout:    layout,
			SourceDir: sourceDir,
		})
		if err != nil {
			return fmt.Errorf("prepare templ css static source: %w", err)
		}
		buildSourceDir = stageDir
		cleanupSource = cleanup
		templCSSStylesheetBuilt = filesystem.PathExists(
			filepath.Join(stageDir, filepath.FromSlash("styles/templ.css")),
		)
	}
	if !layout.ServerFeatures.StaticAssets && !hasClientAssets && !templCSSStylesheetBuilt {
		return cleanStaticAssetOutput(layout)
	}
	if hasClientAssets {
		stageDir, cleanup, err := clientassets.PrepareStaticSource(clientassets.PrepareStaticSourceConfig{
			Layout:    layout,
			SourceDir: buildSourceDir,
		})
		if err != nil {
			return fmt.Errorf("prepare client assets static source: %w", err)
		}
		previousCleanup := cleanupSource
		buildSourceDir = stageDir
		cleanupSource = func() error {
			if cleanupErr := cleanup(); cleanupErr != nil {
				return cleanupErr
			}
			return previousCleanup()
		}
	}

	bundle, err := bundlerstaticassets.Build(bundlerstaticassets.BuildConfig{
		SourceDir: buildSourceDir,
	})
	if err != nil {
		return fmt.Errorf("build static bundle: %w", err)
	}
	defer func() {
		_ = bundle.Cleanup()
	}()

	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("clean output dir %q: %w", outDir, err)
	}
	if err := filesystem.CopyTree(bundle.Dir(), outDir); err != nil {
		return fmt.Errorf("copy processed assets to %q: %w", outDir, err)
	}
	if err := bundlerstaticassets.WriteManifest(manifestPath, bundle.Manifest()); err != nil {
		return fmt.Errorf("write manifest %q: %w", manifestPath, err)
	}
	return nil
}

func cleanStaticAssetOutput(layout projectlayout.ProjectLayout) error {
	outDir := strings.TrimSpace(layout.StaticAssets.OutDir)
	if outDir != "" {
		if err := os.RemoveAll(outDir); err != nil {
			return fmt.Errorf("clean output dir %q: %w", outDir, err)
		}
	}

	manifestPath := strings.TrimSpace(layout.StaticAssets.ManifestPath)
	if manifestPath != "" {
		if err := os.Remove(manifestPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove manifest %q: %w", manifestPath, err)
		}
	}

	return nil
}

func generateTemplInDir(baseDir string, dir string) error {
	hasTempl, err := directoryHasTemplFiles(dir)
	if err != nil {
		return err
	}
	if !hasTempl {
		return nil
	}
	return bundlertemplgen.Run(bundlertemplgen.Config{
		Paths:    []string{dir},
		BasePath: baseDir,
	})
}

func directoryHasTemplFiles(root string) (bool, error) {
	trimmed := strings.TrimSpace(root)
	if trimmed == "" {
		return false, nil
	}
	info, err := os.Stat(trimmed)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !info.IsDir() {
		return false, nil
	}

	hasTempl := false
	walkErr := filepath.WalkDir(trimmed, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(filePath) != ".templ" {
			return nil
		}
		hasTempl = true
		return fs.SkipAll
	})
	if walkErr != nil && !errors.Is(walkErr, fs.SkipAll) {
		return false, walkErr
	}
	return hasTempl, nil
}

func checkGitDiff(rootDir string) error {
	cmd := exec.Command("git", "-C", rootDir, "diff", "--exit-code")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("generated outputs differ from git state: %w", err)
	}
	return nil
}

func exitf(formatText string, args ...any) {
	fmt.Fprintf(os.Stderr, formatText+"\n", args...)
	os.Exit(1)
}
