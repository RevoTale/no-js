package clientassets

import (
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/RevoTale/no-js/internal/bundler/clientassetext"
	"github.com/RevoTale/no-js/internal/projectlayout"
)

func clientAssetRoots(layout projectlayout.ProjectLayout) []string {
	roots := []string{}
	if strings.TrimSpace(layout.RoutesDir) != "" {
		roots = append(roots, layout.RoutesDir)
	}
	componentRoot := componentsDir(layout)
	if strings.TrimSpace(componentRoot) != "" {
		roots = append(roots, componentRoot)
	}
	return roots
}

func componentsDir(layout projectlayout.ProjectLayout) string {
	if strings.TrimSpace(layout.RoutesDir) == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(layout.RoutesDir), "components")
}

func componentsImportPath(layout projectlayout.ProjectLayout) string {
	routesImport := strings.Trim(strings.TrimSpace(layout.RoutesImport), "/")
	if routesImport == "" {
		return "components"
	}
	return path.Join(path.Dir(routesImport), "components")
}

func logicalCSSAssetPath(layout projectlayout.ProjectLayout, filePath string) (string, error) {
	return logicalAssetPath(layout, filePath, ".css")
}

func logicalScriptAssetPath(layout projectlayout.ProjectLayout, filePath string) (string, error) {
	return logicalAssetPath(layout, filePath, ".js")
}

func logicalAssetPath(layout projectlayout.ProjectLayout, filePath string, outputExt string) (string, error) {
	webRoot := filepath.Dir(layout.RoutesDir)
	relative, err := filepath.Rel(webRoot, filePath)
	if err != nil || strings.HasPrefix(relative, "..") {
		relative, err = filepath.Rel(layout.RootDir, filePath)
		if err != nil {
			return "", err
		}
	}
	relative = filepath.ToSlash(relative)
	ext := path.Ext(relative)
	return strings.TrimSuffix(relative, ext) + outputExt, nil
}

func nonEmptyCSSAssets(files []*cssAsset) []*cssAsset {
	out := make([]*cssAsset, 0, len(files))
	for _, file := range files {
		if file == nil || strings.TrimSpace(file.Transformed) == "" {
			continue
		}
		out = append(out, file)
	}
	return out
}

func sortedCSSAssets(values map[*cssAsset]struct{}) []*cssAsset {
	out := make([]*cssAsset, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Slice(out, func(i int, j int) bool { return out[i].SourcePath < out[j].SourcePath })
	return out
}

func sortedScriptAssets(values map[*scriptAsset]struct{}) []*scriptAsset {
	out := make([]*scriptAsset, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Slice(out, func(i int, j int) bool { return out[i].SourcePath < out[j].SourcePath })
	return out
}

func assetStem(filePath string) string {
	return strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
}

func pascalIdentifier(value string) string {
	builder := strings.Builder{}
	capitalize := true
	for _, r := range value {
		if r == '-' || r == '_' || r == '.' || r == ' ' {
			capitalize = true
			continue
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			capitalize = true
			continue
		}
		if builder.Len() == 0 && unicode.IsDigit(r) {
			builder.WriteString("X")
		}
		if capitalize {
			builder.WriteRune(unicode.ToUpper(r))
			capitalize = false
		} else {
			builder.WriteRune(r)
		}
	}
	if builder.Len() == 0 {
		return "Asset"
	}
	return builder.String()
}

func lowerFirst(value string) string {
	if value == "" {
		return value
	}
	r, size := utf8.DecodeRuneInString(value)
	if r == utf8.RuneError && size == 0 {
		return value
	}
	return string(unicode.ToLower(r)) + value[size:]
}

func shouldSkipDir(dir string, layout projectlayout.ProjectLayout) bool {
	clean := filepath.Clean(dir)
	for _, skip := range []string{layout.GeneratedDir, layout.StaticAssets.OutDir, layout.StaticAssets.SourceDir} {
		if strings.TrimSpace(skip) == "" {
			continue
		}
		if clean == filepath.Clean(skip) {
			return true
		}
	}
	base := filepath.Base(clean)
	return strings.HasPrefix(base, ".") || base == "node_modules"
}

func isCSSFile(filePath string) bool {
	return clientassetext.IsCSSFile(filePath)
}

func isScriptFile(filePath string) bool {
	return clientassetext.IsScriptFile(filePath)
}

func isGeneratedHelper(filePath string) bool {
	return clientassetext.IsGeneratedHelperName(filepath.Base(filePath))
}

func isSkipAll(err error) bool {
	return err == fs.SkipAll
}
