package projectlayout

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigMissingFileUsesDefaults(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	cfg, configPath, err := LoadConfig(rootDir)
	require.NoError(t, err)
	require.Equal(t, Config{}, cfg)
	require.Empty(t, configPath)
}

func TestLoadConfigFileRejectsMissingVersion(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "no-js.bundle.yaml")
	writeBundlerTestFile(t, configPath, "project:\n  public_dir: web/public\n")

	_, err := LoadConfigFile(configPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "must declare version: 1")
}

func TestLoadConfigFileRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "no-js.bundle.yaml")
	writeBundlerTestFile(t, configPath, "version: 1\nunknown: value\n")

	_, err := LoadConfigFile(configPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "field unknown not found")
}

func TestLoadConfigFileRejectsInvalidFeatureMode(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "no-js.bundle.yaml")
	writeBundlerTestFile(t, configPath, "version: 1\nserver:\n  features:\n    static_assets: sometimes\n")

	_, err := LoadConfigFile(configPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid feature mode")
}

func TestResolveProjectLayoutFromRootUsesConfigOverrides(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "src", "web", "routes", "page.templ"), "package appsrc\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "src", "web", "view", "context.go"), "package runtime\n")
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
			"  routes_dir: src/web/routes",
			"  generated_dir: src/web/generated",
			"  resolvers_dir: src/web/resolvers",
			"  view_dir: src/web/view",
			"  bootstrap_dir: src/web/bootstrap",
			"  i18n_dir: src/web/i18n",
			"  assets_dir: web-static",
			"  assets_build_dir: web-static-build",
			"  public_dir: web-public",
			"static_assets:",
			"  manifest_path: web-static-build/manifest.json",
			"public_files:",
			"  request_path_prefix: site",
			"",
		}, "\n"),
	)

	layout, err := ResolveProjectLayoutFromRoot(rootDir)
	require.NoError(t, err)

	expectedConfigPath := filepath.ToSlash(filepath.Join(rootDir, defaultBundleConfigFileName))
	require.Equal(t, expectedConfigPath, filepath.ToSlash(layout.ConfigPath))
	require.Equal(t, filepath.ToSlash(filepath.Join(rootDir, "src", "web", "routes")), filepath.ToSlash(layout.RoutesDir))
	expectedBootstrapDir := filepath.ToSlash(filepath.Join(rootDir, "src", "web", "bootstrap"))
	require.Equal(t, expectedBootstrapDir, filepath.ToSlash(layout.BootstrapDir))
	require.Equal(t, filepath.ToSlash(filepath.Join(rootDir, "web-public")), filepath.ToSlash(layout.PublicDir))
	expectedStaticSourceDir := filepath.ToSlash(filepath.Join(rootDir, "web-static"))
	require.Equal(t, expectedStaticSourceDir, filepath.ToSlash(layout.StaticAssets.SourceDir))
	expectedStaticOutDir := filepath.ToSlash(filepath.Join(rootDir, "web-static-build"))
	require.Equal(t, expectedStaticOutDir, filepath.ToSlash(layout.StaticAssets.OutDir))
	require.Equal(t, "/site", layout.PublicRequestPathPrefix)
}
