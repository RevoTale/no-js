package bundler

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigMissingFileUsesDefaults(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	cfg, configPath, err := LoadConfig(rootDir)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg != (Config{}) {
		t.Fatalf("expected zero config when file is missing, got %#v", cfg)
	}
	if configPath != "" {
		t.Fatalf("expected empty config path, got %q", configPath)
	}
}

func TestLoadConfigFileRejectsMissingVersion(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "no-js.bundle.yaml")
	writeBundlerTestFile(t, configPath, "project:\n  public_dir: internal/web/public\n")

	_, err := LoadConfigFile(configPath)
	if err == nil {
		t.Fatal("expected version error")
	}
	if !strings.Contains(err.Error(), "must declare version: 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadConfigFileRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "no-js.bundle.yaml")
	writeBundlerTestFile(t, configPath, "version: 1\nunknown: value\n")

	_, err := LoadConfigFile(configPath)
	if err == nil {
		t.Fatal("expected unknown field error")
	}
	if !strings.Contains(err.Error(), "field unknown not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadConfigFileRejectsInvalidFeatureMode(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "no-js.bundle.yaml")
	writeBundlerTestFile(t, configPath, "version: 1\nserver:\n  features:\n    static_assets: sometimes\n")

	_, err := LoadConfigFile(configPath)
	if err == nil {
		t.Fatal("expected invalid feature mode error")
	}
	if !strings.Contains(err.Error(), "invalid feature mode") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveProjectLayoutFromRootUsesConfigOverrides(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "src", "web", "app", "page.templ"), "package appsrc\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "src", "web", "runtime", "context.go"), "package runtime\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "src", "web", "bootstrap", "bootstrap.go"), "package bootstrap\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "src", "web", "i18n", "doc.go"), "package i18n\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web-public", "robots.txt"), "User-agent: *\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web-static", "app.js"), "console.log('x')\n")
	writeBundlerTestFile(
		t,
		filepath.Join(rootDir, defaultBundleConfigFileName),
		strings.Join([]string{
			"version: 1",
			"project:",
			"  app_dir: src/web/app",
			"  gen_dir: src/web/gen",
			"  resolver_dir: src/web/resolvers",
			"  runtime_dir: src/web/runtime",
			"  bootstrap_dir: src/web/bootstrap",
			"  i18n_dir: src/web/i18n",
			"  public_dir: web-public",
			"static_assets:",
			"  source_dir: web-static",
			"public_files:",
			"  request_path_prefix: site",
			"",
		}, "\n"),
	)

	layout, err := ResolveProjectLayoutFromRoot(rootDir)
	if err != nil {
		t.Fatalf("resolve project layout from root: %v", err)
	}

	expectedConfigPath := filepath.ToSlash(filepath.Join(rootDir, defaultBundleConfigFileName))
	if got := filepath.ToSlash(layout.ConfigPath); got != expectedConfigPath {
		t.Fatalf("unexpected config path: %q", got)
	}
	if got := filepath.ToSlash(layout.AppDir); got != filepath.ToSlash(filepath.Join(rootDir, "src", "web", "app")) {
		t.Fatalf("unexpected app dir: %q", got)
	}
	expectedBootstrapDir := filepath.ToSlash(filepath.Join(rootDir, "src", "web", "bootstrap"))
	if got := filepath.ToSlash(layout.BootstrapDir); got != expectedBootstrapDir {
		t.Fatalf("unexpected bootstrap dir: %q", got)
	}
	if got := filepath.ToSlash(layout.PublicDir); got != filepath.ToSlash(filepath.Join(rootDir, "web-public")) {
		t.Fatalf("unexpected public dir: %q", got)
	}
	expectedStaticSourceDir := filepath.ToSlash(filepath.Join(rootDir, "web-static"))
	if got := filepath.ToSlash(layout.StaticAssets.SourceDir); got != expectedStaticSourceDir {
		t.Fatalf("unexpected static source dir: %q", got)
	}
	if layout.PublicRequestPathPrefix != "/site" {
		t.Fatalf("unexpected public request prefix: %q", layout.PublicRequestPathPrefix)
	}
}
