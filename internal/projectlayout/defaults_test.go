package projectlayout

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
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "routes", "page.templ"), "package appsrc\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "view", "context.go"), "package runtime\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "i18n", "doc.go"), "package i18n\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "i18n", "messages", "active.en.json"), "[]\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "assets", "app.js"), "console.log('x')\n")

	layout, err := ResolveProjectLayout(rootDir, Config{})
	require.NoError(t, err)

	require.Equal(t, filepath.Clean(rootDir), layout.RootDir)
	require.Equal(t, filepath.ToSlash(filepath.Join(rootDir, defaultRoutesDir)), filepath.ToSlash(layout.RoutesDir))
	require.Equal(t, filepath.ToSlash(filepath.Join(rootDir, defaultViewDir)), filepath.ToSlash(layout.ViewDir))
	require.Equal(t, defaultViewDir, layout.ViewImport)
	require.Equal(t, filepath.ToSlash(filepath.Join(rootDir, defaultI18nDir)), filepath.ToSlash(layout.I18nDir))
	require.Equal(t, defaultGeneratedDir, layout.GeneratedImport)
	require.Equal(t, "example.com/app", layout.AppModulePath)
	require.True(t, layout.BuiltInI18n.Enabled)
	expectedMessagesDir := filepath.ToSlash(filepath.Join(rootDir, defaultI18nDir, defaultI18nMessagesDir))
	require.Equal(t, expectedMessagesDir, layout.BuiltInI18n.MessagesDir)
	expectedStaticSourceDir := filepath.ToSlash(filepath.Join(rootDir, defaultAssetsDir))
	require.Equal(t, expectedStaticSourceDir, filepath.ToSlash(layout.StaticAssets.SourceDir))
	expectedStaticOutDir := filepath.ToSlash(filepath.Join(rootDir, defaultAssetsBuildDir))
	require.Equal(t, expectedStaticOutDir, filepath.ToSlash(layout.StaticAssets.OutDir))
	expectedManifestPath := filepath.ToSlash(
		filepath.Join(rootDir, defaultAssetsBuildDir, defaultStaticManifestFileName),
	)
	require.Equal(t, expectedManifestPath, filepath.ToSlash(layout.StaticAssets.ManifestPath))
	require.True(t, layout.ServerFeatures.I18nRouting)
	require.True(t, layout.ServerFeatures.StaticAssets)
	require.True(t, layout.ServerFeatures.HealthEndpoint)
}

func TestResolveProjectLayoutRejectsMissingAppDir(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")

	_, err := ResolveProjectLayout(rootDir, Config{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "strict routes root missing")
}

func TestResolveProjectLayoutRejectsEscapeDir(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "routes", "page.templ"), "package appsrc\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "view", "context.go"), "package runtime\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "i18n", "doc.go"), "package i18n\n")

	_, err := ResolveProjectLayout(rootDir, Config{
		Project: ProjectConfig{
			RoutesDir: "../outside",
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "must stay inside the app root")
}

func TestResolveProjectLayoutDisablesPresenceBasedFeaturesWhenPathsAreMissing(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "routes", "page.templ"), "package appsrc\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "view", "context.go"), "package runtime\n")

	layout, err := ResolveProjectLayout(rootDir, Config{})
	require.NoError(t, err)

	require.False(t, layout.ServerFeatures.I18nRouting)
	require.False(t, layout.BuiltInI18n.Enabled)
	require.False(t, layout.ServerFeatures.StaticAssets)
	require.True(t, layout.ServerFeatures.HealthEndpoint)
}

func TestResolveProjectLayoutRespectsExplicitFeatureOverrides(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "routes", "page.templ"), "package appsrc\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "view", "context.go"), "package runtime\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "i18n", "doc.go"), "package i18n\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "i18n", "messages", "active.en.json"), "[]\n")

	layout, err := ResolveProjectLayout(rootDir, Config{
		Server: ServerConfig{
			Features: ServerFeaturesConfig{
				I18nRouting:    FeatureDisabled,
				HealthEndpoint: FeatureDisabled,
			},
		},
		I18n: BuiltInI18nConfig{
			Mode: FeatureDisabled,
		},
	})
	require.NoError(t, err)

	require.False(t, layout.ServerFeatures.I18nRouting)
	require.False(t, layout.BuiltInI18n.Enabled)
	require.False(t, layout.ServerFeatures.HealthEndpoint)
}

func TestResolveProjectLayoutRequiresI18nDirWhenFeatureEnabled(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "routes", "page.templ"), "package appsrc\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "view", "context.go"), "package runtime\n")

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

func TestResolveProjectLayoutRequiresMessagesDirWhenBuiltInI18nEnabled(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeBundlerTestFile(t, filepath.Join(rootDir, "go.mod"), "module example.com/app\n\ngo 1.25.0\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "routes", "page.templ"), "package appsrc\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "view", "context.go"), "package runtime\n")
	writeBundlerTestFile(t, filepath.Join(rootDir, "web", "i18n", "doc.go"), "package i18n\n")

	_, err := ResolveProjectLayout(rootDir, Config{
		I18n: BuiltInI18nConfig{Mode: FeatureEnabled},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "strict i18n messages root missing")
}

func writeBundlerTestFile(t *testing.T, filePath string, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))
}
