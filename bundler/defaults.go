package bundler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/RevoTale/no-js/internal/filesystem"
	"golang.org/x/mod/modfile"
)

type appRootDir string
func (d appRootDir) moduleDir(moduleName string) string {
	return filepath.ToSlash(filepath.Join(string(d), moduleName))
}

func (d appRootDir) readModulePath() (string, error) {
	goModPath := filepath.Join(string(d), "go.mod")
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


func getAppRoot() (appRootDir, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	if !filesystem.PathExists(filepath.Join(currentDir, defaultAppRouteModule)) {
		return "", fmt.Errorf("strict app root missing: expected %s", defaultAppRouteModule)
	}
	if !filesystem.PathExists(filepath.Join(currentDir, "go.mod")) {
		return "", fmt.Errorf("go.mod is missing in the %s", currentDir)
	}
	return appRootDir(currentDir), nil

}

func ResolvePaths() ( GenerationPaths, error) {
	moduleRoot, err := getAppRoot()
	if err != nil {
		return GenerationPaths{}, err
	}
	appModulePath, err := moduleRoot.readModulePath()
	if err != nil {
		return GenerationPaths{}, err
	}

	return GenerationPaths{
		AppRoot:  moduleRoot.moduleDir(defaultAppRouteModule),
		GenRoot:  moduleRoot.moduleDir(defaultGenModule)   ,
		GenImportRoot: defaultGenModule,
		ResolverRoot:  moduleRoot.moduleDir(defaultResolverModule)   ,
		AppModulePath: appModulePath,
	}, nil
}