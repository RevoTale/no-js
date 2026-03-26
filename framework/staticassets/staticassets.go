package staticassets

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const defaultURLPrefix = "/_assets/"
const CurrentManifestVersion = 1

type Manifest struct {
	Version   int    `json:"version,omitempty"`
	Hash      string `json:"hash"`
	URLPrefix string `json:"url_prefix,omitempty"`
}

func (manifest Manifest) URL(assetPath string) string {
	trimmed := strings.TrimSpace(assetPath)
	trimmed = strings.ReplaceAll(trimmed, `\`, `/`)
	trimmed = strings.TrimPrefix(trimmed, "/")
	return manifest.VersionedURLPrefix(defaultURLPrefix) + trimmed
}

func (manifest Manifest) VersionedURLPrefix(basePrefix string) string {
	normalized := normalizeManifest(manifest)
	return joinVersionedPrefix(basePrefix, normalized.Hash)
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

func NormalizeURLPrefix(prefix string) string {
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

func joinVersionedPrefix(basePrefix string, hash string) string {
	normalizedPrefix := NormalizeURLPrefix(basePrefix)
	trimmedHash := strings.Trim(strings.TrimSpace(hash), "/")
	if trimmedHash == "" {
		return normalizedPrefix
	}

	return normalizedPrefix + trimmedHash + "/"
}

func normalizeManifest(manifest Manifest) Manifest {
	manifest.Hash = strings.TrimSpace(manifest.Hash)
	manifest.URLPrefix = strings.TrimSpace(manifest.URLPrefix)
	if manifest.Hash == "" && manifest.URLPrefix != "" {
		manifest.Hash = hashFromURLPrefix(manifest.URLPrefix)
	}
	if manifest.Version == 0 && manifest.Hash != "" {
		manifest.Version = CurrentManifestVersion
	}
	return manifest
}

func validateManifest(manifest Manifest) error {
	if manifest.Version != CurrentManifestVersion {
		return fmt.Errorf("manifest version must be %d", CurrentManifestVersion)
	}
	if strings.TrimSpace(manifest.Hash) == "" {
		return fmt.Errorf("manifest hash is required")
	}
	return nil
}

func hashFromURLPrefix(prefix string) string {
	normalized := NormalizeURLPrefix(prefix)
	trimmed := strings.Trim(normalized, "/")
	if trimmed == "" {
		return ""
	}

	segments := strings.Split(trimmed, "/")
	if len(segments) == 0 {
		return ""
	}
	return strings.TrimSpace(segments[len(segments)-1])
}
