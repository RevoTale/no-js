package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/RevoTale/no-js/bundler"
	"github.com/RevoTale/no-js/bundler/approutegen"
	bundlerstaticassets "github.com/RevoTale/no-js/bundler/staticassets"
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

func resolveLayout(rootDir string, configPath string) (bundler.ProjectLayout, error) {
	if strings.TrimSpace(configPath) == "" {
		return bundler.ResolveProjectLayoutFromRoot(rootDir)
	}

	cfg, err := bundler.LoadConfigFile(configPath)
	if err != nil {
		return bundler.ProjectLayout{}, err
	}
	layout, err := bundler.ResolveProjectLayout(rootDir, cfg)
	if err != nil {
		return bundler.ProjectLayout{}, err
	}
	resolvedConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return bundler.ProjectLayout{}, fmt.Errorf("resolve config path %q: %w", configPath, err)
	}
	layout.ConfigPath = resolvedConfigPath
	return layout, nil
}

func generateRoutes(layout bundler.ProjectLayout) error {
	if err := approutegen.Run(approutegen.Config{Layout: layout}); err != nil {
		return fmt.Errorf("generate routes: %w", err)
	}
	return nil
}

func generateAssets(layout bundler.ProjectLayout) error {
	if !layout.ServerFeatures.StaticAssets {
		return nil
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

	bundle, err := bundlerstaticassets.Build(bundlerstaticassets.BuildConfig{
		SourceDir: sourceDir,
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
	if err := copyTree(bundle.Dir(), outDir); err != nil {
		return fmt.Errorf("copy processed assets to %q: %w", outDir, err)
	}
	if err := bundlerstaticassets.WriteManifest(manifestPath, bundle.Manifest()); err != nil {
		return fmt.Errorf("write manifest %q: %w", manifestPath, err)
	}
	return nil
}

func copyTree(sourceRoot string, targetRoot string) error {
	return filepath.WalkDir(sourceRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relativePath, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		if relativePath == "." {
			return os.MkdirAll(targetRoot, 0o755)
		}

		targetPath := filepath.Join(targetRoot, relativePath)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(targetPath, content, 0o644); err != nil {
			return err
		}

		return nil
	})
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

func exitf(formatText string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, formatText+"\n", args...)
	os.Exit(1)
}
