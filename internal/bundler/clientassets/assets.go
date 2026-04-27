package clientassets

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/RevoTale/no-js/internal/projectlayout"
)

var packageRE = regexp.MustCompile(`(?m)^\s*package\s+([A-Za-z_][A-Za-z0-9_]*)`)

func discoverAssets(layout projectlayout.ProjectLayout) (*assetIndex, error) {
	index := &assetIndex{
		cssByDir:    map[string][]*cssAsset{},
		scriptByDir: map[string][]*scriptAsset{},
	}
	classAllocator := newGeneratedClassAllocator()
	for _, root := range clientAssetRoots(layout) {
		if _, err := os.Stat(root); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("stat client asset root %q: %w", root, err)
		}
		if err := filepath.WalkDir(root, func(filePath string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if shouldSkipDir(filePath, layout) && filePath != root {
					return fs.SkipDir
				}
				return nil
			}
			if isGeneratedHelper(filePath) {
				return nil
			}
			switch {
			case isCSSFile(filePath):
				asset, err := newCSSAsset(layout, filePath, classAllocator)
				if err != nil {
					return err
				}
				index.cssByDir[asset.PackageDir] = append(index.cssByDir[asset.PackageDir], asset)
				index.cssFiles = append(index.cssFiles, asset)
			case isScriptFile(filePath):
				asset, err := newScriptAsset(layout, filePath)
				if err != nil {
					return err
				}
				index.scriptByDir[asset.PackageDir] = append(index.scriptByDir[asset.PackageDir], asset)
				index.scriptFiles = append(index.scriptFiles, asset)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	sort.Slice(index.cssFiles, func(i int, j int) bool {
		return index.cssFiles[i].SourcePath < index.cssFiles[j].SourcePath
	})
	sort.Slice(index.scriptFiles, func(i int, j int) bool {
		return index.scriptFiles[i].SourcePath < index.scriptFiles[j].SourcePath
	})
	return index, nil
}

func newCSSAsset(
	layout projectlayout.ProjectLayout,
	filePath string,
	classAllocator *generatedClassAllocator,
) (*cssAsset, error) {
	packageName, err := packageNameForDir(filepath.Dir(filePath))
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read css %q: %w", filePath, err)
	}
	classes, transformed, err := transformCSSModule(layout, filePath, string(content), classAllocator)
	if err != nil {
		return nil, err
	}
	assetPath, err := logicalCSSAssetPath(layout, filePath)
	if err != nil {
		return nil, err
	}
	return &cssAsset{
		SourcePath:  filePath,
		PackageDir:  filepath.Dir(filePath),
		PackageName: packageName,
		HelperPath:  strings.TrimSuffix(filePath, ".css") + ".css_gen.go",
		AssetPath:   assetPath,
		Transformed: transformed,
		Classes:     classes,
	}, nil
}

func newScriptAsset(layout projectlayout.ProjectLayout, filePath string) (*scriptAsset, error) {
	packageName, err := packageNameForDir(filepath.Dir(filePath))
	if err != nil {
		return nil, err
	}
	baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	assetPath, err := logicalScriptAssetPath(layout, filePath)
	if err != nil {
		return nil, err
	}
	return &scriptAsset{
		SourcePath:  filePath,
		PackageDir:  filepath.Dir(filePath),
		PackageName: packageName,
		HelperPath:  strings.TrimSuffix(filePath, filepath.Ext(filePath)) + filepath.Ext(filePath) + "_gen.go",
		AssetPath:   assetPath,
		FuncName:    pascalIdentifier(baseName) + "Script",
		VarName:     lowerFirst(pascalIdentifier(baseName)) + "ScriptOnce",
	}, nil
}

func packageNameForDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read package dir %q: %w", dir, err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".templ" {
			files = append(files, filepath.Join(dir, name))
		}
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") || isGeneratedHelper(name) {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	for _, filePath := range files {
		content, readErr := os.ReadFile(filePath)
		if readErr != nil {
			return "", readErr
		}
		match := packageRE.FindSubmatch(content)
		if len(match) >= 2 {
			return string(match[1]), nil
		}
	}
	return "", fmt.Errorf("no Go package declaration found in %q", dir)
}
