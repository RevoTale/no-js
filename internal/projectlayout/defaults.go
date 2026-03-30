package projectlayout

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

	routesDir, routesImportPath, err := resolveModuleDir(resolvedRoot, cfg.Project.RoutesDir, defaultRoutesDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve routes dir: %w", err)
	}
	if !filesystem.PathExists(routesDir) {
		return ProjectLayout{}, fmt.Errorf("strict routes root missing: expected %s", routesImportPath)
	}

	generatedDir, generatedImport, err := resolveModuleDir(resolvedRoot, cfg.Project.GeneratedDir, defaultGeneratedDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve generated dir: %w", err)
	}
	resolversDir, resolversImport, err := resolveModuleDir(resolvedRoot, cfg.Project.ResolversDir, defaultResolversDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve resolvers dir: %w", err)
	}
	viewDir, viewImport, err := resolveModuleDir(resolvedRoot, cfg.Project.ViewDir, defaultViewDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve view dir: %w", err)
	}
	if !filesystem.PathExists(viewDir) {
		return ProjectLayout{}, fmt.Errorf("strict view root missing: expected %s", viewImport)
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
	staticSourceDir, _, err := resolveModuleDir(resolvedRoot, cfg.Project.AssetsDir, defaultAssetsDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve assets dir: %w", err)
	}
	staticOutDir, _, err := resolveModuleDir(resolvedRoot, cfg.Project.AssetsBuildDir, defaultAssetsBuildDir)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("resolve assets build dir: %w", err)
	}
	staticManifestDefault := path.Join(
		strings.Trim(strings.TrimSpace(cfg.Project.AssetsBuildDir), "/"),
		defaultStaticManifestFileName,
	)
	if strings.TrimSpace(staticManifestDefault) == defaultStaticManifestFileName {
		staticManifestDefault = path.Join(defaultAssetsBuildDir, defaultStaticManifestFileName)
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

		RoutesDir:       routesDir,
		RoutesImport:    routesImportPath,
		GeneratedDir:    generatedDir,
		GeneratedImport: generatedImport,
		ResolversDir:    resolversDir,
		ResolversImport: resolversImport,
		ViewDir:         viewDir,
		ViewImport:      viewImport,
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
