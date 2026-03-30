package staticassets

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadManifest(t *testing.T) {
	t.Parallel()

	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(
		manifestPath,
		[]byte("{\n  \"version\": 1,\n  \"hash\": \"abc123\"\n}\n"),
		0o644,
	); err != nil {
		require.NoError(t, err)
	}

	actual, err := ReadManifest(manifestPath)
	require.NoError(t, err)
	require.Equal(t, "abc123", actual.Hash)
	require.Equal(t, CurrentManifestVersion, actual.Version)
}

func TestManifestURL(t *testing.T) {
	t.Parallel()

	manifest := Manifest{
		Version: CurrentManifestVersion,
		Hash:    "abc123",
	}

	require.Equal(t, "/_assets/abc123/styles.css", manifest.URL("/styles.css"))
}

func TestManifestVersionedURLPrefixUsesRuntimePrefix(t *testing.T) {
	t.Parallel()

	manifest := Manifest{
		Version: CurrentManifestVersion,
		Hash:    "abc123",
	}

	require.Equal(t, "/cdn/abc123/", manifest.VersionedURLPrefix("/cdn/"))
}

func TestReadManifestAcceptsLegacyURLPrefix(t *testing.T) {
	t.Parallel()

	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(
		manifestPath,
		[]byte("{\n  \"url_prefix\": \"/legacy/abc123/\"\n}\n"),
		0o644,
	); err != nil {
		require.NoError(t, err)
	}

	actual, err := ReadManifest(manifestPath)
	require.NoError(t, err)
	require.Equal(t, "abc123", actual.Hash)
	require.Equal(t, CurrentManifestVersion, actual.Version)
}
