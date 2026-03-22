package bundler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveProjectLayout(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "app", "page.templ"), "package appsrc\n")

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

	_, err := ResolveProjectLayout(rootDir, Config{
		AppDir: "../outside",
	})
	if err == nil {
		t.Fatal("expected relative escape error")
	}
	if !strings.Contains(err.Error(), "must stay inside the app root") {
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
