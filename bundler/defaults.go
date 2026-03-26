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

	appDir, appImportPath, err := resolveModuleDir(resolvedRoot, cfg.Project.AppDir, defaultAppDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve app dir: %w", err)
	}
	if !filesystem.PathExists(appDir) {
		return ProjectLayout{}, fmt.Errorf("strict app root missing: expected %s", appImportPath)
	}

	generatedDir, generatedImport, err := resolveModuleDir(resolvedRoot, cfg.Project.GenDir, defaultGeneratedDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve generated dir: %w", err)
	}
	resolverDir, resolverImport, err := resolveModuleDir(resolvedRoot, cfg.Project.ResolverDir, defaultResolverDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve resolver dir: %w", err)
	}
	runtimeDir, runtimeImport, err := resolveModuleDir(resolvedRoot, cfg.Project.RuntimeDir, defaultRuntimeDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve runtime dir: %w", err)
	}
	if !filesystem.PathExists(runtimeDir) {
		return ProjectLayout{}, fmt.Errorf("strict runtime root missing: expected %s", runtimeImport)
	}
	bootstrapDir, bootstrapImport, err := resolveModuleDir(resolvedRoot, cfg.Project.BootstrapDir, defaultBootstrapDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve bootstrap dir: %w", err)
	}
	i18nDir, i18nImport, err := resolveModuleDir(resolvedRoot, cfg.Project.I18nDir, defaultI18nDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve i18n dir: %w", err)
	}
	publicDir, _, err := resolveModuleDir(resolvedRoot, cfg.Project.PublicDir, defaultPublicDirName)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve public dir: %w", err)
	}
	staticSourceDir, _, err := resolveModuleDir(resolvedRoot, cfg.StaticAssets.SourceDir, defaultStaticSourceDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve static source dir: %w", err)
	}
	staticOutDir, _, err := resolveModuleDir(resolvedRoot, cfg.StaticAssets.OutDir, defaultStaticOutDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve static out dir: %w", err)
	}
	staticManifestDefault := path.Join(
		strings.Trim(strings.TrimSpace(cfg.StaticAssets.OutDir), "/"),
		defaultStaticManifestFileName,
	)
	if strings.TrimSpace(staticManifestDefault) == defaultStaticManifestFileName {
		staticManifestDefault = path.Join(defaultStaticOutDir, defaultStaticManifestFileName)
	}
	staticManifestPath, err := resolveRelativePath(
		resolvedRoot,
		cfg.StaticAssets.ManifestPath,
		staticManifestDefault,
	)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve static manifest path: %w", err)
	}

	serverFeatures := ServerFeatures{
		I18nRouting:    cfg.Server.Features.I18nRouting.Resolve(filesystem.PathExists(i18nDir)),
		StaticAssets:   cfg.Server.Features.StaticAssets.Resolve(filesystem.PathExists(staticSourceDir)),
		PublicFiles:    cfg.Server.Features.PublicFiles.Resolve(filesystem.PathExists(publicDir)),
		HealthEndpoint: cfg.Server.Features.HealthEndpoint.Resolve(true),
	}
	if serverFeatures.I18nRouting && !filesystem.PathExists(i18nDir) {
		return ProjectLayout{}, fmt.Errorf("strict i18n root missing: expected %s", i18nImport)
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
		ResolverImport:  resolverImport,
		RuntimeDir:      runtimeDir,
		RuntimeImport:   runtimeImport,
		BootstrapDir:    bootstrapDir,
		BootstrapImport: bootstrapImport,
		I18nDir:         i18nDir,
		I18nImport:      i18nImport,

		PublicDir:               publicDir,
		PublicRequestPathPrefix: normalizeRequestPathPrefix(cfg.PublicFiles.RequestPathPrefix),
		StaticAssets: StaticAssetsLayout{
			SourceDir:    staticSourceDir,
			OutDir:       staticOutDir,
			ManifestPath: staticManifestPath,
		},
		ServerFeatures: serverFeatures,

		AppModulePath: appModulePath,
	}, nil
}

func resolveModuleDir(rootDir string, value string, defaultValue string) (string, string, error) {
	moduleDir, err := normalizeRelativePath(value, defaultValue)
	if err != nil {
		return "", "", err
	}

	return filepath.ToSlash(filepath.Join(rootDir, moduleDir)), moduleDir, nil
}

func resolveRelativePath(rootDir string, value string, defaultValue string) (string, error) {
	relativePath, err := normalizeRelativePath(value, defaultValue)
	if err != nil {
		return "", err
	}

	return filepath.ToSlash(filepath.Join(rootDir, relativePath)), nil
}

func normalizeRelativePath(value string, defaultValue string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		trimmed = defaultValue
	}

	normalized := filepath.ToSlash(filepath.Clean(trimmed))
	if normalized == "." {
		return "", fmt.Errorf("path cannot be current directory")
	}
	if path.IsAbs(normalized) {
		return "", fmt.Errorf("path %q must be relative", trimmed)
	}
	if strings.HasPrefix(normalized, "../") || normalized == ".." {
		return "", fmt.Errorf("path %q must stay inside the app root", trimmed)
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
