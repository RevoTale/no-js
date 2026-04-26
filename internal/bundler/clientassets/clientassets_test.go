package clientassets

import (
	"os"
	"path/filepath"
	"regexp"
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

func TestGenerateCSSConstantsWithComplexSelectors(t *testing.T) {
	t.Parallel()

	layout := testLayout(t)
	writeFile(t, filepath.Join(layout.RoutesDir, "root.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.css"), `
.card > .title + .meta::before,
.card:is(.featured, .compact) .icon:not(.disabled) {
	content: ".card";
	border-radius: 9px;
}

.card[data-state=".featured"] > .title {
	color: red;
}

@supports selector(.card:has(> .title + .meta)) {
	.card:has(> .title + .meta)::after {
		background: blue;
	}
}

@container (min-width: 40rem) {
	.card:has(> .title + .meta) > .icon::before {
		outline: 1px solid green;
	}
}

@scope (.card) to (.disabled) {
	.icon {
		text-shadow: 0 1px black;
	}
}

@layer clientassets.e2e {
	.card > .icon {
		outline-offset: 2px;
	}
}
`)

	_, err := Generate(Config{Layout: layout})
	require.NoError(t, err)

	classes := generatedCSSClassValues(t, filepath.Join(layout.RoutesDir, "page.css_gen.go"))
	require.Len(t, classes, 7)
	requireUniqueValues(t, classes)

	stageDir, cleanup, err := PrepareStaticSource(PrepareStaticSourceConfig{Layout: layout})
	require.NoError(t, err)
	defer func() { require.NoError(t, cleanup()) }()

	css := readFile(t, filepath.Join(stageDir, "routes", "index.css"))
	require.Contains(
		t,
		css,
		"."+classes["PageCardClass"]+">."+classes["PageTitleClass"]+"+."+classes["PageMetaClass"]+"::before",
	)
	require.Contains(t, css, ":is(."+classes["PageFeaturedClass"]+",."+classes["PageCompactClass"]+")")
	require.Contains(t, css, ":not(."+classes["PageDisabledClass"]+")")
	require.Contains(t, css, `content:".card"`)
	require.Contains(t, css, `[data-state=".featured"]`)
	require.Regexp(
		t,
		regexp.MustCompile(
			`selector\(\.`+regexp.QuoteMeta(classes["PageCardClass"])+`:has\(`+
				`\s*>\s*\.`+regexp.QuoteMeta(classes["PageTitleClass"])+
				`\s*\+\s*\.`+regexp.QuoteMeta(classes["PageMetaClass"])+`\s*\)\)`,
		),
		css,
	)
	require.Regexp(
		t,
		regexp.MustCompile(
			`@container\s*\(min-width:40rem\)\{\.`+regexp.QuoteMeta(classes["PageCardClass"])+
				`:has\(\s*>\s*\.`+regexp.QuoteMeta(classes["PageTitleClass"])+
				`\s*\+\s*\.`+regexp.QuoteMeta(classes["PageMetaClass"])+`\s*\)\s*>\s*\.`+
				regexp.QuoteMeta(classes["PageIconClass"])+`:{1,2}before`,
		),
		css,
	)
	require.Regexp(
		t,
		regexp.MustCompile(
			`@scope\s*\(\.`+regexp.QuoteMeta(classes["PageCardClass"])+`\)\s*to\s*\(\.`+
				regexp.QuoteMeta(classes["PageDisabledClass"])+`\)\s*\{\s*\.`+
				regexp.QuoteMeta(classes["PageIconClass"])+`\s*\{[^}]*text-shadow:\s*0 1px black`,
		),
		css,
	)
	require.Contains(t, css, "@layer clientassets.e2e{."+classes["PageCardClass"]+">."+classes["PageIconClass"])
	require.NotContains(t, css, "@layer clientassets.n_")
	require.NotContains(t, css, ".card>")
	require.NotContains(t, css, ".title")
	require.NotContains(t, css, ".meta")
}

func TestGenerateCSSReportsInvalidCSS(t *testing.T) {
	t.Parallel()

	layout := testLayout(t)
	writeFile(t, filepath.Join(layout.RoutesDir, "root.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.templ"), "package routes\n")
	writeFile(t, filepath.Join(layout.RoutesDir, "page.css"), ".root { color }\n")

	_, err := Generate(Config{Layout: layout})
	require.Error(t, err)
	require.ErrorContains(t, err, "parse css")
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
		filepath.Join(filepath.Dir(layout.RoutesDir), "components", "meter", "meter.tsx"),
		strings.Join([]string{
			"/** @jsx jsx */",
			"const jsx = (tag: string, props: Record<string, string>) => ({ tag, props });",
			`const marker = <span data-client-assets-meter="loaded" />;`,
			"export const value = marker.props['data-client-assets-meter'];",
			"",
		}, "\n"),
	)

	plan, err := Generate(Config{Layout: layout})
	require.NoError(t, err)

	helper := readFile(t, filepath.Join(filepath.Dir(layout.RoutesDir), "components", "meter", "meter.tsx_gen.go"))
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

func generatedCSSClassValues(t *testing.T, filePath string) map[string]string {
	t.Helper()
	content := readFile(t, filePath)
	re := regexp.MustCompile(`(?m)^\s*([A-Za-z0-9_]+Class) = "(n_[^"]+)"$`)
	matches := re.FindAllStringSubmatch(content, -1)
	require.NotEmpty(t, matches)
	values := map[string]string{}
	for _, match := range matches {
		values[match[1]] = match[2]
	}
	return values
}

func requireUniqueValues(t *testing.T, values map[string]string) {
	t.Helper()
	seen := map[string]string{}
	for name, value := range values {
		if existing, ok := seen[value]; ok {
			t.Fatalf("generated CSS class %q is reused by %s and %s", value, existing, name)
		}
		seen[value] = name
	}
}
