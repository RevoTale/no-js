package bundler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func DefaultConfigPath(rootDir string) string {
	trimmedRoot := strings.TrimSpace(rootDir)
	if trimmedRoot == "" {
		trimmedRoot = "."
	}
	return filepath.Join(trimmedRoot, defaultBundleConfigFileName)
}

func LoadConfig(rootDir string) (Config, string, error) {
	configPath := DefaultConfigPath(rootDir)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Config{}, "", nil
	} else if err != nil {
		return Config{}, "", fmt.Errorf("stat bundle config %q: %w", configPath, err)
	}

	cfg, err := LoadConfigFile(configPath)
	if err != nil {
		return Config{}, "", err
	}
	resolvedPath, err := filepath.Abs(configPath)
	if err != nil {
		return Config{}, "", fmt.Errorf("resolve bundle config path %q: %w", configPath, err)
	}
	return cfg, resolvedPath, nil
}

func LoadConfigFile(path string) (Config, error) {
	configPath := strings.TrimSpace(path)
	if configPath == "" {
		return Config{}, fmt.Errorf("bundle config path is required")
	}

	file, err := os.Open(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("open bundle config %q: %w", configPath, err)
	}
	defer func() {
		_ = file.Close()
	}()

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)

	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode bundle config %q: %w", configPath, err)
	}
	if cfg.Version != bundleConfigVersion {
		return Config{}, fmt.Errorf(
			"bundle config %q must declare version: %d",
			configPath,
			bundleConfigVersion,
		)
	}

	return cfg, nil
}

func ResolveProjectLayoutFromRoot(rootDir string) (ProjectLayout, error) {
	cfg, configPath, err := LoadConfig(rootDir)
	if err != nil {
		return ProjectLayout{}, err
	}

	layout, err := ResolveProjectLayout(rootDir, cfg)
	if err != nil {
		return ProjectLayout{}, err
	}
	layout.ConfigPath = configPath
	return layout, nil
}
