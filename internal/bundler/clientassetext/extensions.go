package clientassetext

import (
	"path/filepath"
	"strings"
)

const CSSExtension = ".css"

var scriptExtensions = []string{".js", ".ts", ".tsx", ".mjs", ".mts"}

// AssetExtensions returns every source extension that creates a Client Asset helper.
func AssetExtensions() []string {
	extensions := make([]string, 0, len(scriptExtensions)+1)
	extensions = append(extensions, CSSExtension)
	extensions = append(extensions, scriptExtensions...)
	return extensions
}

func IsCSSFile(filePath string) bool {
	return filepath.Ext(filePath) == CSSExtension
}

func IsScriptFile(filePath string) bool {
	return IsScriptExtension(filepath.Ext(filePath))
}

func IsAssetExtension(extension string) bool {
	if extension == CSSExtension {
		return true
	}
	return IsScriptExtension(extension)
}

func IsScriptExtension(extension string) bool {
	for _, candidate := range scriptExtensions {
		if extension == candidate {
			return true
		}
	}
	return false
}

func IsGeneratedHelperName(name string) bool {
	for _, extension := range AssetExtensions() {
		if strings.HasSuffix(name, extension+"_gen.go") {
			return true
		}
	}
	return false
}

func IsGeneratedHelperFor(stem string, base string) bool {
	for _, extension := range AssetExtensions() {
		if base == stem+extension+"_gen.go" {
			return true
		}
	}
	return false
}

func GeneratedHelperStem(base string) (string, bool) {
	for _, extension := range AssetExtensions() {
		suffix := extension + "_gen.go"
		if strings.HasSuffix(base, suffix) {
			return strings.TrimSuffix(base, suffix), true
		}
	}
	return "", false
}

func ScriptChoices(stem string) string {
	choices := make([]string, 0, len(scriptExtensions))
	for _, extension := range scriptExtensions {
		choices = append(choices, stem+extension)
	}
	return joinWithOr(choices)
}

func ScriptOutputName(stem string) string {
	return stem + ".js"
}

func joinWithOr(values []string) string {
	switch len(values) {
	case 0:
		return ""
	case 1:
		return values[0]
	case 2:
		return values[0] + " or " + values[1]
	default:
		return strings.Join(values[:len(values)-1], ", ") + ", or " + values[len(values)-1]
	}
}
