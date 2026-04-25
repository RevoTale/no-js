package clientassets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RevoTale/no-js/internal/projectlayout"
	"github.com/stretchr/testify/require"
)

func TestGenerateCSSConstantsAndRouteBundle(t *testing.T) {
	t.Parallel()

	layout := testLayout(t)
	writeFile(t, filepath.Join(layout.RoutesDir, "root.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.css"), ".root { color: red; }\n.root.active { color: blue; }\n")

	plan, err := Generate(Config{Layout: layout})
	require.NoError(t, err)

	helper := readFile(t, filepath.Join(layout.RoutesDir, "page.css_gen.go"))
	require.Contains(t, helper, "const (")
	require.Contains(t, helper, "PageActiveClass")
	require.Contains(t, helper, "PageRootClass")
	require.Contains(t, helper, `"n_`)
	require.Equal(t, []string{"routes/index.css"}, plan.RouteAssets[""].Stylesheets)

	stageDir, cleanup, err := PrepareStaticSource(PrepareStaticSourceConfig{Layout: layout})
	require.NoError(t, err)
	defer func() { require.NoError(t, cleanup()) }()

	css := readFile(t, filepath.Join(stageDir, "routes", "index.css"))
	require.Contains(t, css, ".n_")
	require.NotContains(t, css, ".root")
	require.NotContains(t, css, ".active")
}

func TestGenerateScriptHelpersAndRouteStaticInjectionPlan(t *testing.T) {
	t.Parallel()

	layout := testLayout(t)
	writeFile(t, filepath.Join(layout.RoutesDir, "root.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.templ"), `package routes

import "example.com/client-assets/web/components/meter"
`)
	writeFile(t, filepath.Join(layout.RoutesDir, "about", "page.templ"), "package about\n")
	writeFile(t, filepath.Join(filepath.Dir(layout.RoutesDir), "components", "meter", "meter.templ"), "package meter\n")
	writeFile(
		t,
		filepath.Join(filepath.Dir(layout.RoutesDir), "components", "meter", "meter.ts"),
		"export const value = 1;\n",
	)

	plan, err := Generate(Config{Layout: layout})
	require.NoError(t, err)

	helper := readFile(t, filepath.Join(filepath.Dir(layout.RoutesDir), "components", "meter", "meter.ts_gen.go"))
	require.Contains(t, helper, "func MeterScript() templ.Component")
	require.Contains(t, helper, `metagen.AssetURL(ctx, "components/meter/meter.js")`)
	require.Equal(t, []string{"routes/index.js"}, plan.RouteAssets[""].ModuleScripts)
	require.Empty(t, plan.RouteAssets["about"].ModuleScripts)

	stageDir, cleanup, err := PrepareStaticSource(PrepareStaticSourceConfig{Layout: layout})
	require.NoError(t, err)
	defer func() { require.NoError(t, cleanup()) }()

	require.FileExists(t, filepath.Join(stageDir, "components", "meter", "meter.js"))
	require.FileExists(t, filepath.Join(stageDir, "routes", "index.js"))
	require.NoFileExists(t, filepath.Join(stageDir, "routes", "about.js"))
}

func testLayout(t *testing.T) projectlayout.ProjectLayout {
	t.Helper()
	root := t.TempDir()
	return projectlayout.ProjectLayout{
		RootDir:       root,
		RoutesDir:     filepath.Join(root, "web", "routes"),
		RoutesImport:  "web/routes",
		GeneratedDir:  filepath.Join(root, "web", "generated"),
		AppModulePath: "example.com/client-assets",
		StaticAssets: projectlayout.StaticAssetsLayout{
			SourceDir:    filepath.Join(root, "web", "assets"),
			OutDir:       filepath.Join(root, "web", "assets-build"),
			ManifestPath: filepath.Join(root, "web", "assets-build", "manifest.json"),
		},
	}
}

func writeFile(t *testing.T, filePath string, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))
}

func readFile(t *testing.T, filePath string) string {
	t.Helper()
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	return strings.TrimSpace(string(content))
}
