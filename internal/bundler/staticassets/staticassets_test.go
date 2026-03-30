package staticassets

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	frameworkstaticassets "github.com/RevoTale/no-js/framework/staticassets"
	"github.com/stretchr/testify/require"
)

func TestBuild_MinifiesAndCopiesAssets(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()
	writeTestFile(
		t,
		filepath.Join(sourceDir, "app.js"),
		"function sum (a, b) {\n  return a + b\n}\nconsole.log(sum(1, 2))\n",
	)
	writeTestFile(t, filepath.Join(sourceDir, "styles.css"), "body {\n  color: red;\n  margin: 0;\n}\n")
	writeTestFile(t, filepath.Join(sourceDir, "logo.svg"), "<svg>\n  <rect width=\"10\" height=\"10\"/>\n</svg>\n")

	bundle, err := Build(BuildConfig{
		SourceDir: sourceDir,
		URLPrefix: "/_assets/",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, bundle.Cleanup())
	})

	prefixPattern := regexp.MustCompile(`^/_assets/[0-9a-f]{16}/$`)
	require.True(t, prefixPattern.MatchString(bundle.URLPrefix()))
	require.Equal(t, bundle.URLPrefix()+"styles.css", bundle.URL("styles.css"))

	minifiedJS := mustReadFile(t, filepath.Join(bundle.Dir(), "app.js"))
	originalJS := mustReadFile(t, filepath.Join(sourceDir, "app.js"))
	require.NotEqual(t, string(originalJS), string(minifiedJS))
	require.Less(t, len(minifiedJS), len(originalJS))

	minifiedCSS := mustReadFile(t, filepath.Join(bundle.Dir(), "styles.css"))
	originalCSS := mustReadFile(t, filepath.Join(sourceDir, "styles.css"))
	require.NotEqual(t, string(originalCSS), string(minifiedCSS))
	require.Less(t, len(minifiedCSS), len(originalCSS))

	copiedSVG := mustReadFile(t, filepath.Join(bundle.Dir(), "logo.svg"))
	originalSVG := mustReadFile(t, filepath.Join(sourceDir, "logo.svg"))
	require.Equal(t, string(originalSVG), string(copiedSVG))
}

func TestBuild_HashDeterministicForSameSource(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()
	writeTestFile(t, filepath.Join(sourceDir, "app.js"), "const value = 1 + 2;\nconsole.log(value)\n")
	writeTestFile(t, filepath.Join(sourceDir, "styles.css"), "body { color: blue; }\n")
	writeTestFile(t, filepath.Join(sourceDir, "logo.svg"), "<svg><circle cx=\"5\" cy=\"5\" r=\"2\" /></svg>\n")

	first, err := Build(BuildConfig{SourceDir: sourceDir, URLPrefix: "/_assets/"})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, first.Cleanup())
	})

	second, err := Build(BuildConfig{SourceDir: sourceDir, URLPrefix: "/_assets/"})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, second.Cleanup())
	})

	require.Equal(t, first.Hash(), second.Hash())
}

func TestBuild_HashChangesWhenContentChanges(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()
	writeTestFile(t, filepath.Join(sourceDir, "app.js"), "console.log('a')\n")

	first, err := Build(BuildConfig{SourceDir: sourceDir, URLPrefix: "/_assets/"})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, first.Cleanup())
	})

	writeTestFile(t, filepath.Join(sourceDir, "app.js"), "console.log('b')\n")
	second, err := Build(BuildConfig{SourceDir: sourceDir, URLPrefix: "/_assets/"})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, second.Cleanup())
	})

	require.NotEqual(t, first.Hash(), second.Hash())
}

func TestBuild_HashChangesWhenRelativePathChanges(t *testing.T) {
	t.Parallel()

	firstDir := t.TempDir()
	secondDir := t.TempDir()

	writeTestFile(t, filepath.Join(firstDir, "a.js"), "console.log('same')\n")
	writeTestFile(t, filepath.Join(secondDir, "nested", "a.js"), "console.log('same')\n")

	first, err := Build(BuildConfig{SourceDir: firstDir, URLPrefix: "/_assets/"})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, first.Cleanup())
	})

	second, err := Build(BuildConfig{SourceDir: secondDir, URLPrefix: "/_assets/"})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, second.Cleanup())
	})

	require.NotEqual(t, first.Hash(), second.Hash())
}

func TestBundleCleanupRemovesOutputDir(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()
	writeTestFile(t, filepath.Join(sourceDir, "app.js"), "console.log('x')\n")

	bundle, err := Build(BuildConfig{SourceDir: sourceDir, URLPrefix: "/_assets/"})
	require.NoError(t, err)

	outDir := bundle.Dir()
	require.NotEmpty(t, outDir)
	_, statErr := os.Stat(outDir)
	require.NoError(t, statErr)

	require.NoError(t, bundle.Cleanup())
	_, statErr = os.Stat(outDir)
	require.True(t, os.IsNotExist(statErr))
}

func TestWriteManifest(t *testing.T) {
	t.Parallel()

	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	expected := frameworkstaticassets.Manifest{
		Version: frameworkstaticassets.CurrentManifestVersion,
		Hash:    "abc123",
	}

	require.NoError(t, WriteManifest(manifestPath, expected))

	actual, err := frameworkstaticassets.ReadManifest(manifestPath)
	require.NoError(t, err)
	require.Equal(t, expected.Hash, actual.Hash)
	require.Equal(t, expected.Version, actual.Version)
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	return content
}
