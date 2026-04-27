package clientassets

import (
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/RevoTale/no-js/internal/projectlayout"
)

type dependencyCollector struct {
	layout       projectlayout.ProjectLayout
	index        *assetIndex
	visitedFiles map[string]struct{}
	visitedDirs  map[string]struct{}
	css          map[*cssAsset]struct{}
	scripts      map[*scriptAsset]struct{}
}

func newDependencyCollector(layout projectlayout.ProjectLayout, index *assetIndex) *dependencyCollector {
	return &dependencyCollector{
		layout:       layout,
		index:        index,
		visitedFiles: map[string]struct{}{},
		visitedDirs:  map[string]struct{}{},
		css:          map[*cssAsset]struct{}{},
		scripts:      map[*scriptAsset]struct{}{},
	}
}

func (collector *dependencyCollector) addFile(filePath string) {
	cleanPath := filepath.Clean(filePath)
	if _, ok := collector.visitedFiles[cleanPath]; ok {
		return
	}
	collector.visitedFiles[cleanPath] = struct{}{}
	collector.addFileAssets(cleanPath)

	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return
	}
	for _, importPath := range importPathsForFile(cleanPath, string(content)) {
		if dir, ok := collector.componentDirForImport(importPath); ok {
			collector.addDir(dir)
			collector.scanPackageFiles(dir)
		}
	}
}

func (collector *dependencyCollector) addFileAssets(filePath string) {
	stem := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	if stem == "" {
		return
	}

	dir := filepath.Dir(filepath.Clean(filePath))
	for _, asset := range collector.index.cssByDir[dir] {
		if assetStem(asset.SourcePath) == stem {
			collector.css[asset] = struct{}{}
		}
	}
	if stem == "root" {
		return
	}
	for _, asset := range collector.index.scriptByDir[dir] {
		if assetStem(asset.SourcePath) == stem {
			collector.scripts[asset] = struct{}{}
		}
	}
}

func (collector *dependencyCollector) addDir(dir string) {
	cleanDir := filepath.Clean(dir)
	if _, ok := collector.visitedDirs[cleanDir]; ok {
		return
	}
	collector.visitedDirs[cleanDir] = struct{}{}
	for _, asset := range collector.index.cssByDir[cleanDir] {
		collector.css[asset] = struct{}{}
	}
	for _, asset := range collector.index.scriptByDir[cleanDir] {
		collector.scripts[asset] = struct{}{}
	}
}

func (collector *dependencyCollector) scanPackageFiles(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".templ" && filepath.Ext(name) != ".go" {
			continue
		}
		if strings.HasSuffix(name, "_test.go") || isGeneratedHelper(name) {
			continue
		}
		collector.addFile(filepath.Join(dir, name))
	}
}

func (collector *dependencyCollector) componentDirForImport(importPath string) (string, bool) {
	prefix := path.Join(collector.layout.AppModulePath, componentsImportPath(collector.layout))
	trimmed := strings.TrimSpace(importPath)
	if trimmed != prefix && !strings.HasPrefix(trimmed, prefix+"/") {
		return "", false
	}
	relative := strings.TrimPrefix(trimmed, prefix)
	relative = strings.TrimPrefix(relative, "/")
	return filepath.Join(componentsDir(collector.layout), filepath.FromSlash(relative)), true
}

func importPathsForFile(filePath string, content string) []string {
	file, err := parser.ParseFile(token.NewFileSet(), filePath, content, parser.ImportsOnly)
	if err != nil {
		return nil
	}
	paths := make([]string, 0, len(file.Imports))
	for _, spec := range file.Imports {
		if spec.Path == nil {
			continue
		}
		value, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		paths = append(paths, value)
	}
	return paths
}
