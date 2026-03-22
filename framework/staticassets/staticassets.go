package staticassets

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const defaultURLPrefix = "/_assets/"

type Manifest struct {
	Hash      string `json:"hash"`
	URLPrefix string `json:"url_prefix"`
}

func (manifest Manifest) URL(assetPath string) string {
	manifest = normalizeManifest(manifest)

	trimmed := strings.TrimSpace(assetPath)
	trimmed = strings.ReplaceAll(trimmed, `\`, `/`)
	trimmed = strings.TrimPrefix(trimmed, "/")
	return manifest.URLPrefix + trimmed
}

func ReadManifest(path string) (Manifest, error) {
	manifestPath := strings.TrimSpace(path)
	if manifestPath == "" {
		return Manifest{}, fmt.Errorf("manifest path is required")
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest %q: %w", manifestPath, err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest %q: %w", manifestPath, err)
	}

	manifest = normalizeManifest(manifest)
	if err := validateManifest(manifest); err != nil {
		return Manifest{}, err
	}

	return manifest, nil
}

func normalizeURLPrefix(prefix string) string {
	trimmed := strings.TrimSpace(prefix)
	if trimmed == "" {
		trimmed = defaultURLPrefix
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	if !strings.HasSuffix(trimmed, "/") {
		trimmed += "/"
	}

	return trimmed
}

func normalizeManifest(manifest Manifest) Manifest {
	manifest.Hash = strings.TrimSpace(manifest.Hash)
	manifest.URLPrefix = normalizeURLPrefix(manifest.URLPrefix)
	return manifest
}

func validateManifest(manifest Manifest) error {
	if strings.TrimSpace(manifest.Hash) == "" {
		return fmt.Errorf("manifest hash is required")
	}
	if strings.TrimSpace(manifest.URLPrefix) == "" {
		return fmt.Errorf("manifest url prefix is required")
	}
	return nil
}
