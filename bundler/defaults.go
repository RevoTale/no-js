package bundler

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/RevoTale/no-js/internal/filesystem"
	"golang.org/x/mod/modfile"
)

func ResolveProjectLayout(rootDir string, cfg Config) (ProjectLayout, error) {
	resolvedRoot, err := filepath.Abs(strings.TrimSpace(rootDir))
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve root dir %q: %w", rootDir, err)
	}
	if strings.TrimSpace(resolvedRoot) == "" {
		return ProjectLayout{}, fmt.Errorf("root dir is required")
	}
	if !filesystem.PathExists(filepath.Join(resolvedRoot, "go.mod")) {
		return ProjectLayout{}, fmt.Errorf("go.mod is missing in %s", resolvedRoot)
	}

	appDir, appImportPath, err := resolveModuleDir(resolvedRoot, cfg.AppDir, defaultAppDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve app dir: %w", err)
	}
	if !filesystem.PathExists(appDir) {
		return ProjectLayout{}, fmt.Errorf("strict app root missing: expected %s", appImportPath)
	}

	generatedDir, generatedImport, err := resolveModuleDir(resolvedRoot, cfg.GenDir, defaultGeneratedDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve generated dir: %w", err)
	}
	resolverDir, _, err := resolveModuleDir(resolvedRoot, cfg.Resolver, defaultResolverDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve resolver dir: %w", err)
	}
	publicDir, _, err := resolveModuleDir(resolvedRoot, cfg.PublicDirName, defaultPublicDirName)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve public dir: %w", err)
	}

	appModulePath, err := readModulePath(resolvedRoot)
	if err != nil {
		return ProjectLayout{}, err
	}

	return ProjectLayout{
		RootDir: resolvedRoot,

		AppDir:          appDir,
		GeneratedDir:    generatedDir,
		GeneratedImport: generatedImport,
		ResolverDir:     resolverDir,

		PublicDir:               publicDir,
		PublicRequestPathPrefix: normalizeRequestPathPrefix(cfg.PublicDirRequestPathPrefix),

		AppModulePath: appModulePath,
	}, nil
}

func resolveModuleDir(rootDir string, value string, defaultValue string) (string, string, error) {
	moduleDir, err := normalizeModuleDir(value, defaultValue)
	if err != nil {
		return "", "", err
	}

	return filepath.ToSlash(filepath.Join(rootDir, moduleDir)), moduleDir, nil
}

func normalizeModuleDir(value string, defaultValue string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		trimmed = defaultValue
	}

	normalized := filepath.ToSlash(filepath.Clean(trimmed))
	if normalized == "." {
		return "", fmt.Errorf("module dir cannot be current directory")
	}
	if path.IsAbs(normalized) {
		return "", fmt.Errorf("module dir %q must be relative", trimmed)
	}
	if strings.HasPrefix(normalized, "../") || normalized == ".." {
		return "", fmt.Errorf("module dir %q must stay inside the app root", trimmed)
	}

	return normalized, nil
}

func normalizeRequestPathPrefix(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return defaultPublicRequestPathPrefix
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	if trimmed != "/" && strings.HasSuffix(trimmed, "/") {
		trimmed = strings.TrimRight(trimmed, "/")
	}
	return trimmed
}

func readModulePath(rootDir string) (string, error) {
	goModPath := filepath.Join(rootDir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("read %q: %w", filepath.ToSlash(goModPath), err)
	}

	modulePath := strings.TrimSpace(modfile.ModulePath(data))
	if modulePath == "" {
		return "", fmt.Errorf("module path missing in %q", filepath.ToSlash(goModPath))
	}

	return modulePath, nil
}
