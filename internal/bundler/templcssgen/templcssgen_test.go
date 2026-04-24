package templcssgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RevoTale/no-js/internal/projectlayout"
	"github.com/stretchr/testify/require"
)

func TestRunGeneratesWorkspaceAndRegistryWithoutSourcePollution(t *testing.T) {
	t.Parallel()

	layout := newTemplCSSLayout(t)

	require.NoError(t, os.MkdirAll(filepath.Join(layout.RoutesDir, "notes"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(layout.RootDir, "web/components/card"), 0o755))
	require.NoError(t, os.MkdirAll(layout.ViewDir, 0o755))
	require.NoError(t, os.MkdirAll(layout.GeneratedDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(layout.RootDir, "go.mod"), []byte("module example.com/app\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.ViewDir, "variants.go"), []byte("package view\n"), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(layout.RoutesDir, "notes", "page.templ"), []byte(`
package notes

css button() {
	color: white;
}

css loading(percent int) {
	width: { percent };
}
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.RootDir, "web/components/card", "card.templ"), []byte(`
package card

css panel() {
	border: 1px solid black;
}
`), 0o644))

	require.NoError(t, Run(Config{Layout: layout}))

	_, err := os.Stat(filepath.Join(layout.RoutesDir, "notes", packageExportFileName))
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Join(layout.RootDir, "web/components/card", packageExportFileName))
	require.ErrorIs(t, err, os.ErrNotExist)

	routeExportPath := filepath.Join(
		layout.GeneratedDir,
		generatedWorkspaceDirName,
		"routes",
		"notes",
		packageExportFileName,
	)
	routeExport, err := os.ReadFile(routeExportPath)
	require.NoError(t, err)
	require.Contains(t, string(routeExport), "func "+packageExportFuncName+"() []templ.CSSClass")
	require.Contains(t, string(routeExport), "button()")
	require.NotContains(t, string(routeExport), "loading()")

	componentExportPath := filepath.Join(
		layout.GeneratedDir,
		generatedWorkspaceDirName,
		"components",
		"card",
		packageExportFileName,
	)
	componentExport, err := os.ReadFile(componentExportPath)
	require.NoError(t, err)
	require.Contains(t, string(componentExport), "panel()")

	_, err = os.Stat(filepath.Join(layout.GeneratedDir, generatedWorkspaceDirName, "routes", "notes", "page.templ"))
	require.NoError(t, err)

	registry, err := os.ReadFile(filepath.Join(layout.GeneratedDir, registryFileName))
	require.NoError(t, err)
	require.NotContains(t, string(registry), `"example.com/app/web/view"`)
	require.NotContains(t, string(registry), `view.TemplCSSVariants()`)
	require.Contains(t, string(registry), `example.com/app/web/generated/templcss/components/card`)
	require.Contains(t, string(registry), `example.com/app/web/generated/templcss/routes/notes`)
}

func TestRunIncludesTemplCSSVariantsHookWhenPresent(t *testing.T) {
	t.Parallel()

	layout := newTemplCSSLayout(t)

	require.NoError(t, os.MkdirAll(filepath.Join(layout.RoutesDir, "notes"), 0o755))
	require.NoError(t, os.MkdirAll(layout.ViewDir, 0o755))
	require.NoError(t, os.MkdirAll(layout.GeneratedDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(layout.RootDir, "go.mod"), []byte("module example.com/app\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.ViewDir, "variants.go"), []byte(`
package view

import "github.com/a-h/templ"

func TemplCSSVariants() []templ.CSSClass { return nil }
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.RoutesDir, "notes", "page.templ"), []byte(`
package notes

css button() {
	color: white;
}
`), 0o644))

	require.NoError(t, Run(Config{Layout: layout}))

	registry, err := os.ReadFile(filepath.Join(layout.GeneratedDir, registryFileName))
	require.NoError(t, err)
	require.Contains(t, string(registry), `"example.com/app/web/view"`)
	require.NotContains(t, string(registry), `view "example.com/app/web/view"`)
	require.Contains(t, string(registry), `view.TemplCSSVariants()`)
}

func TestRunRemovesStalePackageExport(t *testing.T) {
	t.Parallel()

	layout := newTemplCSSLayout(t)
	require.NoError(t, os.MkdirAll(filepath.Join(layout.RoutesDir, "notes"), 0o755))
	require.NoError(t, os.MkdirAll(layout.ViewDir, 0o755))
	staleFile := filepath.Join(layout.RoutesDir, "notes", packageExportFileName)
	require.NoError(t, os.WriteFile(staleFile, []byte("stale"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.RootDir, "go.mod"), []byte("module example.com/app\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.RoutesDir, "notes", "page.templ"), []byte(`
package notes

templ Page() {}
`), 0o644))

	require.NoError(t, Run(Config{Layout: layout}))

	_, err := os.Stat(staleFile)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func newTemplCSSLayout(t *testing.T) projectlayout.ProjectLayout {
	t.Helper()

	root := t.TempDir()
	return projectlayout.ProjectLayout{
		RootDir:         filepath.ToSlash(root),
		RoutesDir:       filepath.ToSlash(filepath.Join(root, "web/routes")),
		GeneratedDir:    filepath.ToSlash(filepath.Join(root, "web/generated")),
		ViewDir:         filepath.ToSlash(filepath.Join(root, "web/view")),
		ViewImport:      "web/view",
		GeneratedImport: "web/generated",
		AppModulePath:   "example.com/app",
	}
}
