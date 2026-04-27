package appshape

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RevoTale/no-js/internal/projectlayout"
	"github.com/stretchr/testify/require"
)

func TestValidateAllowsRouteAndComponentGenerationInputs(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)

	writeFile(t, filepath.Join(layout.RoutesDir, "root.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "404.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "root.css"), ":root {}\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.css"), ".shell {}\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.ts"), "console.log('page')\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.css_gen.go"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "section", "layout.templ"), "package section\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "section", "layout.mts"), "console.log('layout')\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "section", "route.go"), "package section\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "robots.go"), "package routes\n")

	componentDir := filepath.Join(root, "web", "components", "card")
	writeFile(t, filepath.Join(componentDir, "card.templ"), "package card\n")
	writeFile(t, filepath.Join(componentDir, "card.go"), "package card\n\nfunc CardModel() string { return \"card\" }\n")
	writeFile(t, filepath.Join(componentDir, "helpers.go"), strings.Join([]string{
		"package card",
		"",
		"func cardLabel() string {",
		"	return \"card\"",
		"}",
		"",
	}, "\n"))
	writeFile(t, filepath.Join(componentDir, "card_test.go"), "package card\n\nfunc TestCard() {}\n")
	writeFile(t, filepath.Join(componentDir, "card.css"), ".card {}\n")
	writeFile(t, filepath.Join(componentDir, "card.ts"), "console.log('card')\n")
	writeFile(t, filepath.Join(componentDir, "card_templ.go"), "package card\n")
	writeFile(t, filepath.Join(componentDir, "card.css_gen.go"), "package card\n")
	writeFile(t, filepath.Join(componentDir, "templ_css_exports_gen.go"), "package card\n")

	require.NoError(t, Validate(layout))
}

func TestValidateAllowsCodeOnlyComponentAnchor(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	componentDir := filepath.Join(root, "web", "components", "badge")
	writeFile(t, filepath.Join(componentDir, "badge.go"), "package badge\n\ntype Badge struct{}\n")
	writeFile(t, filepath.Join(componentDir, "helpers.go"), "package badge\n\nfunc helper() {}\n")

	require.NoError(t, Validate(layout))
}

func TestValidateAllowsGeneratedAndTestComponentFiles(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	componentDir := filepath.Join(root, "web", "components", "card")
	writeFile(t, filepath.Join(componentDir, "card.templ"), "package card\n")
	writeFile(t, filepath.Join(componentDir, "card_test.go"), "package card_test\n\nfunc Helper() {}\n")
	writeFile(t, filepath.Join(componentDir, "card_templ.go"), "package other\n\nfunc Helper() {}\n")
	writeFile(t, filepath.Join(componentDir, "card.css_gen.go"), "package other\n\nfunc Helper() {}\n")
	writeFile(t, filepath.Join(componentDir, "templ_css_exports_gen.go"), "package other\n\nfunc Helper() {}\n")

	require.NoError(t, Validate(layout))
}

func TestValidateRejectsUnsupportedRouteFile(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	writeFile(t, filepath.Join(layout.RoutesDir, "root.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "helper.go"), "package routes\n")

	err := Validate(layout)
	require.Error(t, err)
	require.Contains(t, err.Error(), `unsupported file in web/routes: "helper.go"`)
}

func TestValidateRejectsRootScriptAsset(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	writeFile(t, filepath.Join(layout.RoutesDir, "root.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "root.ts"), "console.log('root')\n")

	err := Validate(layout)
	require.Error(t, err)
	require.Contains(t, err.Error(), `unsupported file in web/routes: "root.ts"`)
}

func TestValidateRejectsRouteAssetWithoutMatchingTemplate(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	writeFile(t, filepath.Join(layout.RoutesDir, "root.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "dashboard", "page.css"), ".shell {}\n")

	err := Validate(layout)
	require.Error(t, err)
	require.Contains(
		t,
		err.Error(),
		`route Client Asset "dashboard/page.css" requires matching template "dashboard/page.templ"`,
	)
}

func TestValidateRejectsMultipleRouteScriptSources(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	writeFile(t, filepath.Join(layout.RoutesDir, "root.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.ts"), "console.log('page')\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.tsx"), "console.log('page')\n")

	err := Validate(layout)
	require.Error(t, err)
	require.Contains(t, err.Error(), `route directory web/routes has multiple script sources for "page.templ"`)
	require.Contains(t, err.Error(), `choose one of page.js, page.ts, page.tsx, page.mjs, or page.mts`)
}

func TestValidateRejectsMixedCaseRouteAssetExtension(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	writeFile(t, filepath.Join(layout.RoutesDir, "root.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.CSS"), ".shell {}\n")

	err := Validate(layout)
	require.Error(t, err)
	require.Contains(t, err.Error(), `unsupported file in web/routes: "page.CSS"`)
}

func TestValidateRejectsLooseRootComponentFile(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	writeFile(t, filepath.Join(root, "web", "components", "badge.templ"), "package components\n")

	err := Validate(layout)
	require.Error(t, err)
	require.Contains(t, err.Error(), `unsupported file in web/components: "badge.templ"`)
	require.Contains(t, err.Error(), `web/components/badge/badge.templ`)
}

func TestValidateRejectsComponentPackageMissingAnchor(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	writeFile(t, filepath.Join(root, "web", "components", "theme", "helpers.go"), "package theme\n\nfunc helper() {}\n")
	writeFile(t, filepath.Join(root, "web", "components", "theme", "theme.css"), ".theme {}\n")

	err := Validate(layout)
	require.Error(t, err)
	require.Contains(t, err.Error(), `component package web/components/theme must contain "theme.templ" or "theme.go"`)
}

func TestValidateRejectsExtraComponentTemplate(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	writeFile(t, filepath.Join(root, "web", "components", "card", "card.templ"), "package card\n")
	writeFile(t, filepath.Join(root, "web", "components", "card", "variants.templ"), "package card\n")

	err := Validate(layout)
	require.Error(t, err)
	require.Contains(t, err.Error(), `unsupported component template "card/variants.templ"`)
	require.Contains(t, err.Error(), `card.templ`)
}

func TestValidateRejectsNonSameStemComponentAsset(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	writeFile(t, filepath.Join(root, "web", "components", "card", "card.templ"), "package card\n")
	writeFile(t, filepath.Join(root, "web", "components", "card", "theme.css"), ".theme {}\n")

	err := Validate(layout)
	require.Error(t, err)
	require.Contains(t, err.Error(), `unsupported component asset "card/theme.css"`)
	require.Contains(t, err.Error(), `web/components/card/card.css`)
}

func TestValidateRejectsMultipleComponentScriptSources(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	writeFile(t, filepath.Join(root, "web", "components", "card", "card.templ"), "package card\n")
	writeFile(t, filepath.Join(root, "web", "components", "card", "card.ts"), "console.log('card')\n")
	writeFile(t, filepath.Join(root, "web", "components", "card", "card.tsx"), "console.log('card')\n")

	err := Validate(layout)
	require.Error(t, err)
	require.Contains(t, err.Error(), `component package web/components/card has multiple script sources`)
	require.Contains(t, err.Error(), `choose one of card.js, card.ts, card.tsx, card.mjs, or card.mts`)
}

func TestValidateRejectsMixedCaseComponentAssetExtension(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	writeFile(t, filepath.Join(root, "web", "components", "card", "card.templ"), "package card\n")
	writeFile(t, filepath.Join(root, "web", "components", "card", "card.TSX"), "console.log('card')\n")

	err := Validate(layout)
	require.Error(t, err)
	require.Contains(t, err.Error(), `unsupported file in web/components: "card/card.TSX"`)
}

func TestValidateRejectsUnsupportedComponentFile(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	writeFile(t, filepath.Join(root, "web", "components", "card", "card.templ"), "package card\n")
	writeFile(t, filepath.Join(root, "web", "components", "card", "logo.svg"), "<svg/>\n")

	err := Validate(layout)
	require.Error(t, err)
	require.Contains(t, err.Error(), `unsupported file in web/components: "card/logo.svg"`)
}

func TestValidateRejectsPackageNameMismatch(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	writeFile(t, filepath.Join(root, "web", "components", "card", "card.templ"), "package components\n")

	err := Validate(layout)
	require.Error(t, err)
	require.Contains(t, err.Error(), `component template "card/card.templ" declares package "components"`)
	require.Contains(t, err.Error(), `must use package card`)
}

func TestValidateRejectsExportedDeclarationsFromSupportGo(t *testing.T) {
	tests := map[string]string{
		"function": "package card\n\nfunc Helper() {}\n",
		"method":   "package card\n\ntype local struct{}\n\nfunc (local) Helper() {}\n",
		"type":     "package card\n\ntype Helper struct{}\n",
		"var":      "package card\n\nvar Helper = \"helper\"\n",
		"const":    "package card\n\nconst Helper = \"helper\"\n",
	}

	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			layout := testLayout(root)
			writeFile(t, filepath.Join(root, "web", "components", "card", "card.templ"), "package card\n")
			writeFile(t, filepath.Join(root, "web", "components", "card", "helpers.go"), source)

			err := Validate(layout)
			require.Error(t, err)
			require.Contains(t, err.Error(), `exported declaration "Helper" in web/components/card/helpers.go`)
			require.Contains(t, err.Error(), `must move to "card.go" or be made private`)
		})
	}
}

func TestValidateAllowsExportedDeclarationsInAnchorGo(t *testing.T) {
	root := t.TempDir()
	layout := testLayout(root)
	writeFile(t, filepath.Join(root, "web", "components", "card", "card.go"), strings.Join([]string{
		"package card",
		"",
		"func Helper() {}",
		"type Model struct{}",
		"var Value = \"value\"",
		"const Name = \"card\"",
		"",
	}, "\n"))
	writeFile(t, filepath.Join(root, "web", "components", "card", "helpers.go"), "package card\n\nfunc helper() {}\n")

	require.NoError(t, Validate(layout))
}

func testLayout(root string) projectlayout.ProjectLayout {
	return projectlayout.ProjectLayout{
		RoutesDir: filepath.Join(root, "web", "routes"),
	}
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
}
