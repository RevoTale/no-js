package templcssgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RevoTale/no-js/internal/projectlayout"
	"github.com/stretchr/testify/require"
)

func TestRunGeneratesPackageExportsAndRegistry(t *testing.T) {
	t.Parallel()

	layout := newTemplCSSLayout(t)

	require.NoError(t, os.MkdirAll(filepath.Join(layout.RoutesDir, "notes"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(layout.RootDir, "web/components/card"), 0o755))
	require.NoError(t, os.MkdirAll(layout.ViewDir, 0o755))
	require.NoError(t, os.MkdirAll(layout.GeneratedDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(layout.RootDir, "go.mod"), []byte("module example.com/app\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.ViewDir, "variants.go"), []byte("package runtime\n"), 0o644))

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

	routeExport, err := os.ReadFile(filepath.Join(layout.RoutesDir, "notes", packageExportFileName))
	require.NoError(t, err)
	require.Contains(t, string(routeExport), "func "+packageExportFuncName+"() []templ.CSSClass")
	require.Contains(t, string(routeExport), "button()")
	require.NotContains(t, string(routeExport), "loading()")

	componentExport, err := os.ReadFile(filepath.Join(layout.RootDir, "web/components/card", packageExportFileName))
	require.NoError(t, err)
	require.Contains(t, string(componentExport), "panel()")

	registry, err := os.ReadFile(filepath.Join(layout.GeneratedDir, registryFileName))
	require.NoError(t, err)
	require.Contains(t, string(registry), `runtime "example.com/app/web/view"`)
	require.Contains(t, string(registry), `runtime.TemplCSSVariants()`)
	require.Contains(t, string(registry), `example.com/app/web/components/card`)
	require.Contains(t, string(registry), `example.com/app/web/routes/notes`)
}

func TestRunRemovesStalePackageExport(t *testing.T) {
	t.Parallel()

	layout := newTemplCSSLayout(t)
	require.NoError(t, os.MkdirAll(filepath.Join(layout.RoutesDir, "notes"), 0o755))
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
