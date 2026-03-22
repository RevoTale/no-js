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
		[]byte("{\n  \"hash\": \"abc123\",\n  \"url_prefix\": \"/_assets/abc123/\"\n}\n"),
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
	if actual.URLPrefix != "/_assets/abc123/" {
		t.Fatalf("prefix mismatch: got %q", actual.URLPrefix)
	}
}

func TestManifestURL(t *testing.T) {
	t.Parallel()

	manifest := Manifest{
		Hash:      "abc123",
		URLPrefix: "/_assets/abc123/",
	}

	if got := manifest.URL("/styles.css"); got != "/_assets/abc123/styles.css" {
		t.Fatalf("unexpected asset url: %q", got)
	}
}
