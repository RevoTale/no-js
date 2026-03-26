package bundler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveProjectLayoutDefaults(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "app", "page.templ"), "package appsrc\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "runtime", "context.go"), "package runtime\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "i18n", "doc.go"), "package i18n\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "static", "app.js"), "console.log('x')\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "public", "robots.txt"), "User-agent: *\n")

	layout, err := ResolveProjectLayout(rootDir, Config{})
	if err != nil {
		t.Fatalf("resolve project layout: %v", err)
	}

	if layout.RootDir != filepath.Clean(rootDir) {
		t.Fatalf("unexpected root dir: %q", layout.RootDir)
	}
	if got := filepath.ToSlash(layout.AppDir); got != filepath.ToSlash(filepath.Join(rootDir, defaultAppDir)) {
		t.Fatalf("unexpected app dir: %q", got)
	}
	if got := filepath.ToSlash(layout.RuntimeDir); got != filepath.ToSlash(filepath.Join(rootDir, defaultRuntimeDir)) {
		t.Fatalf("unexpected runtime dir: %q", got)
	}
	if layout.RuntimeImport != defaultRuntimeDir {
		t.Fatalf("unexpected runtime import: %q", layout.RuntimeImport)
	}
	if got := filepath.ToSlash(layout.BootstrapDir); got != filepath.ToSlash(filepath.Join(rootDir, defaultBootstrapDir)) {
		t.Fatalf("unexpected bootstrap dir: %q", got)
	}
	if layout.BootstrapImport != defaultBootstrapDir {
		t.Fatalf("unexpected bootstrap import: %q", layout.BootstrapImport)
	}
	if got := filepath.ToSlash(layout.I18nDir); got != filepath.ToSlash(filepath.Join(rootDir, defaultI18nDir)) {
		t.Fatalf("unexpected i18n dir: %q", got)
	}
	if layout.GeneratedImport != defaultGeneratedDir {
		t.Fatalf("unexpected generated import: %q", layout.GeneratedImport)
	}
	if layout.AppModulePath != "example.com/app" {
		t.Fatalf("unexpected app module path: %q", layout.AppModulePath)
	}
	if got := filepath.ToSlash(layout.PublicDir); got != filepath.ToSlash(filepath.Join(rootDir, defaultPublicDirName)) {
		t.Fatalf("unexpected public dir: %q", got)
	}
	if layout.PublicRequestPathPrefix != "/" {
		t.Fatalf("unexpected public request path prefix: %q", layout.PublicRequestPathPrefix)
	}
	expectedStaticSourceDir := filepath.ToSlash(filepath.Join(rootDir, defaultStaticSourceDir))
	if got := filepath.ToSlash(layout.StaticAssets.SourceDir); got != expectedStaticSourceDir {
		t.Fatalf("unexpected static source dir: %q", got)
	}
	expectedStaticOutDir := filepath.ToSlash(filepath.Join(rootDir, defaultStaticOutDir))
	if got := filepath.ToSlash(layout.StaticAssets.OutDir); got != expectedStaticOutDir {
		t.Fatalf("unexpected static out dir: %q", got)
	}
	expectedManifestPath := filepath.ToSlash(
		filepath.Join(rootDir, defaultStaticOutDir, defaultStaticManifestFileName),
	)
	if got := filepath.ToSlash(layout.StaticAssets.ManifestPath); got != expectedManifestPath {
		t.Fatalf("unexpected static manifest path: %q", got)
	}
	if !layout.ServerFeatures.I18nRouting {
		t.Fatalf("expected i18n routing to auto-enable")
	}
	if !layout.ServerFeatures.StaticAssets {
		t.Fatalf("expected static assets to auto-enable")
	}
	if !layout.ServerFeatures.PublicFiles {
		t.Fatalf("expected public files to auto-enable")
	}
	if !layout.ServerFeatures.HealthEndpoint {
		t.Fatalf("expected health endpoint to auto-enable")
	}
}

func TestResolveProjectLayoutRejectsMissingAppDir(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")

	_, err := ResolveProjectLayout(rootDir, Config{})
	if err == nil {
		t.Fatal("expected missing app dir error")
	}
	if !strings.Contains(err.Error(), "strict app root missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveProjectLayoutRejectsEscapeDir(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "app", "page.templ"), "package appsrc\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "runtime", "context.go"), "package runtime\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "i18n", "doc.go"), "package i18n\n")

	_, err := ResolveProjectLayout(rootDir, Config{
		Project: ProjectConfig{
			AppDir: "../outside",
		},
	})
	if err == nil {
		t.Fatal("expected relative escape error")
	}
	if !strings.Contains(err.Error(), "must stay inside the app root") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveProjectLayoutDisablesPresenceBasedFeaturesWhenPathsAreMissing(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "app", "page.templ"), "package appsrc\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "runtime", "context.go"), "package runtime\n")

	layout, err := ResolveProjectLayout(rootDir, Config{})
	if err != nil {
		t.Fatalf("resolve project layout: %v", err)
	}

	if layout.ServerFeatures.I18nRouting {
		t.Fatalf("expected i18n routing to stay disabled when i18n dir is missing")
	}
	if layout.ServerFeatures.StaticAssets {
		t.Fatalf("expected static assets to stay disabled when static dir is missing")
	}
	if layout.ServerFeatures.PublicFiles {
		t.Fatalf("expected public files to stay disabled when public dir is missing")
	}
	if !layout.ServerFeatures.HealthEndpoint {
		t.Fatalf("expected health endpoint to remain enabled")
	}
}

func TestResolveProjectLayoutRespectsExplicitFeatureOverrides(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "app", "page.templ"), "package appsrc\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "runtime", "context.go"), "package runtime\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "i18n", "doc.go"), "package i18n\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "public", "robots.txt"), "User-agent: *\n")

	layout, err := ResolveProjectLayout(rootDir, Config{
		Server: ServerConfig{
			Features: ServerFeaturesConfig{
				I18nRouting:    FeatureDisabled,
				PublicFiles:    FeatureDisabled,
				HealthEndpoint: FeatureDisabled,
			},
		},
		PublicFiles: PublicFilesConfig{
			RequestPathPrefix: "site/",
		},
	})
	if err != nil {
		t.Fatalf("resolve project layout: %v", err)
	}

	if layout.ServerFeatures.I18nRouting {
		t.Fatalf("expected explicit i18n disable to win over presence-based auto")
	}
	if layout.ServerFeatures.PublicFiles {
		t.Fatalf("expected explicit public files disable to win over presence-based auto")
	}
	if layout.ServerFeatures.HealthEndpoint {
		t.Fatalf("expected explicit health disable to win over default auto")
	}
	if layout.PublicRequestPathPrefix != "/site" {
		t.Fatalf("unexpected public request prefix: %q", layout.PublicRequestPathPrefix)
	}
}

func TestResolveProjectLayoutRequiresI18nDirWhenFeatureEnabled(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "app", "page.templ"), "package appsrc\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "runtime", "context.go"), "package runtime\n")

	_, err := ResolveProjectLayout(rootDir, Config{
		Server: ServerConfig{
			Features: ServerFeaturesConfig{
				I18nRouting: FeatureEnabled,
			},
		},
	})
	if err == nil {
		t.Fatal("expected missing i18n dir error")
	}
	if !strings.Contains(err.Error(), "strict i18n root missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeBundlerTestFile(t *testing.T, filePath string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(filePath), err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", filePath, err)
	}
}
