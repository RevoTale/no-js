package bundler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	require.Equal(t, filepath.Clean(rootDir), layout.RootDir)
	require.Equal(t, filepath.ToSlash(filepath.Join(rootDir, defaultAppDir)), filepath.ToSlash(layout.AppDir))
	require.Equal(t, filepath.ToSlash(filepath.Join(rootDir, defaultRuntimeDir)), filepath.ToSlash(layout.RuntimeDir))
	require.Equal(t, defaultRuntimeDir, layout.RuntimeImport)
	require.Equal(t, filepath.ToSlash(filepath.Join(rootDir, defaultBootstrapDir)), filepath.ToSlash(layout.BootstrapDir))
	require.Equal(t, defaultBootstrapDir, layout.BootstrapImport)
	require.Equal(t, filepath.ToSlash(filepath.Join(rootDir, defaultI18nDir)), filepath.ToSlash(layout.I18nDir))
	require.Equal(t, defaultGeneratedDir, layout.GeneratedImport)
	require.Equal(t, "example.com/app", layout.AppModulePath)
	require.Equal(t, filepath.ToSlash(filepath.Join(rootDir, defaultPublicDirName)), filepath.ToSlash(layout.PublicDir))
	require.Equal(t, "/", layout.PublicRequestPathPrefix)
	expectedStaticSourceDir := filepath.ToSlash(filepath.Join(rootDir, defaultStaticSourceDir))
	require.Equal(t, expectedStaticSourceDir, filepath.ToSlash(layout.StaticAssets.SourceDir))
	expectedStaticOutDir := filepath.ToSlash(filepath.Join(rootDir, defaultStaticOutDir))
	require.Equal(t, expectedStaticOutDir, filepath.ToSlash(layout.StaticAssets.OutDir))
	expectedManifestPath := filepath.ToSlash(
		filepath.Join(rootDir, defaultStaticOutDir, defaultStaticManifestFileName),
	)
	require.Equal(t, expectedManifestPath, filepath.ToSlash(layout.StaticAssets.ManifestPath))
	require.True(t, layout.ServerFeatures.I18nRouting)
	require.True(t, layout.ServerFeatures.StaticAssets)
	require.True(t, layout.ServerFeatures.PublicFiles)
	require.True(t, layout.ServerFeatures.HealthEndpoint)
}

func TestResolveProjectLayoutRejectsMissingAppDir(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")

	_, err := ResolveProjectLayout(rootDir, Config{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "strict app root missing")
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
	require.Error(t, err)
	require.Contains(t, err.Error(), "must stay inside the app root")
}

func TestResolveProjectLayoutDisablesPresenceBasedFeaturesWhenPathsAreMissing(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "app", "page.templ"), "package appsrc\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "internal", "web", "runtime", "context.go"), "package runtime\n")

	layout, err := ResolveProjectLayout(rootDir, Config{})
	require.NoError(t, err)

	require.False(t, layout.ServerFeatures.I18nRouting)
	require.False(t, layout.ServerFeatures.StaticAssets)
	require.False(t, layout.ServerFeatures.PublicFiles)
	require.True(t, layout.ServerFeatures.HealthEndpoint)
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
	require.NoError(t, err)

	require.False(t, layout.ServerFeatures.I18nRouting)
	require.False(t, layout.ServerFeatures.PublicFiles)
	require.False(t, layout.ServerFeatures.HealthEndpoint)
	require.Equal(t, "/site", layout.PublicRequestPathPrefix)
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
	require.Error(t, err)
	require.Contains(t, err.Error(), "strict i18n root missing")
}

func writeBundlerTestFile(t *testing.T, filePath string, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))
}
