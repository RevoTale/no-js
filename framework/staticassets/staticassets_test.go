package staticassets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadManifest(t *testing.T) {
	t.Parallel()

	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(
		manifestPath,
		[]byte("{\n  \"version\": 1,\n  \"hash\": \"abc123\"\n}\n"),
		0o644,
	); err != nil {
		t.Fatalf("write manifest fixture: %v", err)
	}

	actual, err := ReadManifest(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	if actual.Hash != "abc123" {
		t.Fatalf("hash mismatch: got %q", actual.Hash)
	}
	if actual.Version != CurrentManifestVersion {
		t.Fatalf("version mismatch: got %d", actual.Version)
	}
}

func TestManifestURL(t *testing.T) {
	t.Parallel()

	manifest := Manifest{
		Version: CurrentManifestVersion,
		Hash:    "abc123",
	}

	if got := manifest.URL("/styles.css"); got != "/_assets/abc123/styles.css" {
		t.Fatalf("unexpected asset url: %q", got)
	}
}

func TestManifestVersionedURLPrefixUsesRuntimePrefix(t *testing.T) {
	t.Parallel()

	manifest := Manifest{
		Version: CurrentManifestVersion,
		Hash:    "abc123",
	}

	if got := manifest.VersionedURLPrefix("/cdn/"); got != "/cdn/abc123/" {
		t.Fatalf("unexpected versioned prefix: %q", got)
	}
}

func TestReadManifestAcceptsLegacyURLPrefix(t *testing.T) {
	t.Parallel()

	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(
		manifestPath,
		[]byte("{\n  \"url_prefix\": \"/legacy/abc123/\"\n}\n"),
		0o644,
	); err != nil {
		t.Fatalf("write manifest fixture: %v", err)
	}

	actual, err := ReadManifest(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	if actual.Hash != "abc123" {
		t.Fatalf("hash mismatch: got %q", actual.Hash)
	}
	if actual.Version != CurrentManifestVersion {
		t.Fatalf("version mismatch: got %d", actual.Version)
	}
}
