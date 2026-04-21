package templcssgen

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	bundlerstaticassets "github.com/RevoTale/no-js/internal/bundler/staticassets"
	bundlertemplgen "github.com/RevoTale/no-js/internal/bundler/templgen"
	"github.com/RevoTale/no-js/internal/projectlayout"
	"github.com/stretchr/testify/require"
)

const templModuleSum = "github.com/a-h/templ v0.3.1001 h1:yHDTgexACdJttyiyamcTHXr2QkIeVF1MukLy44EAhMY=\n" +
	"github.com/a-h/templ v0.3.1001/go.mod h1:oCZcnKRf5jjsGpf2yELzQfodLphd2mwecwG4Crk5HBo=\n"

func tempAppGoMod(repoRoot string) string {
	return "module example.com/app\n\n" +
		"go 1.25.0\n\n" +
		"require (\n" +
		"\tgithub.com/RevoTale/no-js v0.0.0\n" +
		"\tgithub.com/a-h/templ v0.3.1001\n" +
		")\n\n" +
		"replace github.com/RevoTale/no-js => " + filepath.ToSlash(repoRoot) + "\n"
}

func TestWriteStylesheetIncludesZeroArgAndExplicitVariants(t *testing.T) {
	t.Parallel()

	repoRoot := repoRootPath(t)
	appRoot := t.TempDir()
	layout := projectlayout.ProjectLayout{
		RootDir:         filepath.ToSlash(appRoot),
		RoutesDir:       filepath.ToSlash(filepath.Join(appRoot, "web/routes")),
		GeneratedDir:    filepath.ToSlash(filepath.Join(appRoot, "web/generated")),
		GeneratedImport: "web/generated",
		ViewDir:         filepath.ToSlash(filepath.Join(appRoot, "web/view")),
		ViewImport:      "web/view",
		AppModulePath:   "example.com/app",
	}

	require.NoError(t, os.MkdirAll(layout.RoutesDir, 0o755))
	require.NoError(t, os.MkdirAll(layout.ViewDir, 0o755))
	require.NoError(t, os.MkdirAll(layout.GeneratedDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(appRoot, "web/assets"), 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(appRoot, "go.mod"), []byte(tempAppGoMod(repoRoot)), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(appRoot, "go.sum"), []byte(templModuleSum), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(layout.RoutesDir, "page.templ"), []byte(`
package routes

css button() {
	color: white;
}
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.ViewDir, "variants.templ"), []byte(`
package runtime

import "fmt"

css loading(percent int) {
	width: { fmt.Sprintf("%d%%", percent) };
}
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.ViewDir, "variants.go"), []byte(`
package runtime

import "github.com/a-h/templ"

func TemplCSSVariants() []templ.CSSClass {
	return []templ.CSSClass{
		loading(50),
	}
}
`), 0o644))

	require.NoError(t, bundlertemplgen.Run(bundlertemplgen.Config{
		Paths:    []string{layout.ViewDir},
		BasePath: appRoot,
	}))
	require.NoError(t, Run(Config{Layout: layout}))

	outputPath := filepath.Join(appRoot, "web/assets", "styles", "templ.css")
	require.NoError(t, WriteStylesheet(WriteStylesheetConfig{
		Layout:     layout,
		OutputPath: outputPath,
	}))

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	require.Contains(t, string(content), ".button_")
	require.Contains(t, string(content), "color:white")
	require.Contains(t, string(content), "width:50%")
	_, err = os.Stat(filepath.Join(layout.RoutesDir, "page_templ.go"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestWriteStylesheetFromRouteCSSWithoutSourceTemplgen(t *testing.T) {
	t.Parallel()

	repoRoot := repoRootPath(t)
	appRoot := t.TempDir()
	layout := projectlayout.ProjectLayout{
		RootDir:         filepath.ToSlash(appRoot),
		RoutesDir:       filepath.ToSlash(filepath.Join(appRoot, "web/routes")),
		GeneratedDir:    filepath.ToSlash(filepath.Join(appRoot, "web/generated")),
		GeneratedImport: "web/generated",
		ViewDir:         filepath.ToSlash(filepath.Join(appRoot, "web/view")),
		ViewImport:      "web/view",
		AppModulePath:   "example.com/app",
	}

	require.NoError(t, os.MkdirAll(layout.RoutesDir, 0o755))
	require.NoError(t, os.MkdirAll(layout.ViewDir, 0o755))
	require.NoError(t, os.MkdirAll(layout.GeneratedDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(appRoot, "web/assets"), 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(appRoot, "go.mod"), []byte(tempAppGoMod(repoRoot)), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(appRoot, "go.sum"), []byte(templModuleSum), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.ViewDir, "variants.go"), []byte(`
package runtime

import "github.com/a-h/templ"

func TemplCSSVariants() []templ.CSSClass { return nil }
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.RoutesDir, "page.templ"), []byte(`
package routes

css pageShell() {
	padding: 16px;
	border: 1px solid #123;
}
`), 0o644))

	require.NoError(t, Run(Config{Layout: layout}))

	outputPath := filepath.Join(appRoot, "web/assets", "styles", "templ.css")
	require.NoError(t, WriteStylesheet(WriteStylesheetConfig{
		Layout:     layout,
		OutputPath: outputPath,
	}))

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	require.Contains(t, string(content), ".pageShell_")
	require.Contains(t, string(content), "padding:16px")
	require.Contains(t, string(content), "border:1px solid #123")
	_, err = os.Stat(filepath.Join(layout.RoutesDir, "page_templ.go"))
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Join(layout.RoutesDir, packageExportFileName))
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Join(layout.GeneratedDir, generatedWorkspaceDirName, "routes", "page_templ.go"))
	require.NoError(t, err)
}

func TestPrepareStaticSourceStagesTemplCSSForStaticBundle(t *testing.T) {
	t.Parallel()

	repoRoot := repoRootPath(t)
	appRoot := t.TempDir()
	layout := projectlayout.ProjectLayout{
		RootDir:         filepath.ToSlash(appRoot),
		RoutesDir:       filepath.ToSlash(filepath.Join(appRoot, "web/routes")),
		GeneratedDir:    filepath.ToSlash(filepath.Join(appRoot, "web/generated")),
		GeneratedImport: "web/generated",
		ViewDir:         filepath.ToSlash(filepath.Join(appRoot, "web/view")),
		ViewImport:      "web/view",
		AppModulePath:   "example.com/app",
	}

	require.NoError(t, os.MkdirAll(layout.RoutesDir, 0o755))
	require.NoError(t, os.MkdirAll(layout.ViewDir, 0o755))
	require.NoError(t, os.MkdirAll(layout.GeneratedDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(appRoot, "web/assets"), 0o755))
	require.NoError(
		t,
		os.WriteFile(filepath.Join(appRoot, "web/assets", "site.css"), []byte("body { color: black; }\n"), 0o644),
	)

	require.NoError(t, os.WriteFile(filepath.Join(appRoot, "go.mod"), []byte(tempAppGoMod(repoRoot)), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(appRoot, "go.sum"), []byte(templModuleSum), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(layout.RoutesDir, "page.templ"), []byte(`
package routes

css button() {
	color: white;
}
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.ViewDir, "variants.go"), []byte(`
package runtime

import "github.com/a-h/templ"

func TemplCSSVariants() []templ.CSSClass {
	return nil
}
`), 0o644))

	stageDir, cleanup, err := PrepareStaticSource(PrepareStaticSourceConfig{
		Layout:    layout,
		SourceDir: filepath.Join(appRoot, "web/assets"),
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, cleanup())
	}()

	bundle, err := bundlerstaticassets.Build(bundlerstaticassets.BuildConfig{
		SourceDir: stageDir,
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, bundle.Cleanup())
	}()

	content, err := os.ReadFile(filepath.Join(bundle.Dir(), "styles", "templ.css"))
	require.NoError(t, err)
	require.Contains(t, string(content), ".button_")
	require.Contains(t, bundle.URL("styles/templ.css"), "/styles/templ.css")
}

func TestPrepareStaticSourceStagesTemplCSSWithoutSourceDir(t *testing.T) {
	t.Parallel()

	repoRoot := repoRootPath(t)
	appRoot := t.TempDir()
	layout := projectlayout.ProjectLayout{
		RootDir:         filepath.ToSlash(appRoot),
		RoutesDir:       filepath.ToSlash(filepath.Join(appRoot, "web/routes")),
		GeneratedDir:    filepath.ToSlash(filepath.Join(appRoot, "web/generated")),
		GeneratedImport: "web/generated",
		ViewDir:         filepath.ToSlash(filepath.Join(appRoot, "web/view")),
		ViewImport:      "web/view",
		AppModulePath:   "example.com/app",
	}

	require.NoError(t, os.MkdirAll(layout.RoutesDir, 0o755))
	require.NoError(t, os.MkdirAll(layout.ViewDir, 0o755))
	require.NoError(t, os.MkdirAll(layout.GeneratedDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(appRoot, "go.mod"), []byte(tempAppGoMod(repoRoot)), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(appRoot, "go.sum"), []byte(templModuleSum), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(layout.RoutesDir, "page.templ"), []byte(`
package routes

css button() {
	color: white;
}
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.ViewDir, "variants.go"), []byte(`
package runtime

import "github.com/a-h/templ"

func TemplCSSVariants() []templ.CSSClass {
	return nil
}
`), 0o644))

	stageDir, cleanup, err := PrepareStaticSource(PrepareStaticSourceConfig{
		Layout:    layout,
		SourceDir: filepath.Join(appRoot, "web/assets"),
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, cleanup())
	}()

	bundle, err := bundlerstaticassets.Build(bundlerstaticassets.BuildConfig{
		SourceDir: stageDir,
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, bundle.Cleanup())
	}()

	content, err := os.ReadFile(filepath.Join(bundle.Dir(), "styles", "templ.css"))
	require.NoError(t, err)
	require.Contains(t, string(content), ".button_")
	require.Contains(t, bundle.URL("styles/templ.css"), "/styles/templ.css")
}

func TestStylesheetHelperSourceReturnsErrorsInsteadOfPanicking(t *testing.T) {
	t.Parallel()

	source, err := stylesheetHelperSource(projectlayout.ProjectLayout{
		AppModulePath:   "example.com/app",
		GeneratedImport: "web/generated",
	}, "/tmp/styles/templ.css")
	require.NoError(t, err)
	require.NotContains(t, string(source), "panic(")
	require.Contains(t, string(source), "func run() error")
	require.Contains(t, string(source), "fmt.Fprintln(os.Stderr, err)")
}

func repoRootPath(t *testing.T) string {
	t.Helper()

	_, fileName, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(fileName), "..", "..", ".."))
}
