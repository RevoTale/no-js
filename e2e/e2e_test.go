package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoutePageCSSUsesGlobalStylesheetWithoutInlineStyle(t *testing.T) {
	_, report := loadRoutePageCSSFixture(t)

	require.Regexp(
		t,
		regexp.MustCompile(`^/_assets/[0-9a-f]{16}/styles/templ\.css$`),
		report.StylesheetURL,
	)

	require.Equal(t, 200, report.Home.Status)
	require.Contains(t, report.Home.ContentType, "text/html")
	require.Contains(t, report.Home.Body, `data-route-page="home"`)
	require.Contains(t, report.Home.Body, report.StylesheetURL)
	require.NotContains(t, report.Home.Body, `<style type="text/css"`)

	pageClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<main[^>]*data-route-page="home"[^>]*>`, report.Home.Body),
	)
	requireClassWithToken(t, report.Stylesheet.Body, pageClasses, "padding:16px")
	requireClassWithToken(t, report.Stylesheet.Body, pageClasses, "border:1px solid #123")

	require.Equal(t, 200, report.Partial.Status)
	require.NotContains(t, report.Partial.Body, `<html`)
	require.NotContains(t, report.Partial.Body, `<style type="text/css"`)
	require.Contains(t, report.Partial.HXTriggerAfterSettle, `styles/templ.css`)

	require.Equal(t, 200, report.Stylesheet.Status)
	require.Contains(t, report.Stylesheet.ContentType, "text/css")
}

func TestRoutePageCSSWorksViaNoJSGenWithoutSourcePollution(t *testing.T) {
	appDir, _ := loadRoutePageCSSFixture(t)

	_, err := os.Stat(filepath.Join(appDir, "web/routes", "page_templ.go"))
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Join(appDir, "web/routes", "root_templ.go"))
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Join(appDir, "web/routes", "templ_css_exports_gen.go"))
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Join(appDir, "web/generated", "templcss", "routes", "page_templ.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(appDir, "web/generated", "templcss", "routes", "templ_css_exports_gen.go"))
	require.NoError(t, err)
}

func TestFixtureAppsDeclareToolsAndLocalReplaceInGoMod(t *testing.T) {
	t.Helper()

	fixturesDir := filepath.Join(repoRootPath(t), "e2e", "testdata")
	entries, err := os.ReadDir(fixturesDir)
	require.NoError(t, err)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		goModPath := filepath.Join(fixturesDir, entry.Name(), "go.mod")
		goMod, err := os.ReadFile(goModPath)
		require.NoError(t, err, "%s missing go.mod", entry.Name())

		contents := string(goMod)
		require.Contains(t, contents, "module example.com/no-js-e2e/"+entry.Name())
		require.Contains(t, contents, "tool (")
		require.Contains(t, contents, "github.com/RevoTale/no-js/cmd/no-js")
		require.Contains(t, contents, "github.com/RevoTale/no-js/cmd/templgen")
		require.Contains(t, contents, "github.com/evanw/esbuild v0.28.0 // indirect")
		require.Contains(t, contents, "github.com/nicksnyder/go-i18n/v2 v2.6.1 // indirect")
		require.Contains(t, contents, "github.com/tdewolff/parse/v2 v2.8.12 // indirect")
		require.Contains(t, contents, "replace github.com/RevoTale/no-js => ../../..")

		goSumPath := filepath.Join(fixturesDir, entry.Name(), "go.sum")
		goSum, err := os.ReadFile(goSumPath)
		require.NoError(t, err, "%s missing go.sum", entry.Name())
		require.Contains(t, string(goSum), "github.com/evanw/esbuild v0.28.0/go.mod")
		require.Contains(t, string(goSum), "github.com/tdewolff/parse/v2 v2.8.12/go.mod")

		serverPath := filepath.Join(fixturesDir, entry.Name(), "cmd", "server", "main.go")
		_, err = os.Stat(serverPath)
		require.NoError(t, err, "%s missing cmd/server/main.go", entry.Name())

		probePath := filepath.Join(fixturesDir, entry.Name(), "cmd", "probe", "main.go")
		_, err = os.Stat(probePath)
		require.ErrorIs(t, err, os.ErrNotExist, "%s still has cmd/probe/main.go", entry.Name())
	}
}

func TestClientAssetsFixtureApp(t *testing.T) {
	report := loadClientAssetsFixture(t)

	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/index\.css$`), report.RouteCSSURL)
	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/index\.js$`), report.RouteScriptURL)
	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/section\.css$`), report.SectionCSSURL)
	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/section\.js$`), report.SectionScriptURL)
	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/complex\.css$`), report.ComplexCSSURL)
	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/404\.css$`), report.NotFoundCSSURL)

	require.Equal(t, 200, report.Home.Status)
	require.Contains(t, report.Home.Body, `data-route-page="home"`)
	require.Contains(t, report.Home.Body, report.RouteCSSURL)
	require.Contains(t, report.Home.Body, report.RouteScriptURL)
	require.Equal(t, 1, strings.Count(report.Home.Body, report.RouteCSSURL))
	require.Equal(t, 1, strings.Count(report.Home.Body, report.RouteScriptURL))
	require.NotContains(t, report.Home.Body, report.SectionCSSURL)
	require.NotContains(t, report.Home.Body, report.SectionScriptURL)
	require.NotContains(t, report.Home.Body, report.ComplexCSSURL)
	require.NotContains(t, report.Home.Body, "PageShellClass")
	require.NotContains(t, report.Home.Body, "MeterRootClass")
	require.Contains(t, report.Home.Body, `class="n_`)

	shellClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<main[^>]*data-route-page="home"[^>]*>`, report.Home.Body),
	)
	secondItemClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<div[^>]*data-hard-case="second"[^>]*>`, report.Home.Body),
	)
	labelClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<span[^>]*data-hard-case="label"[^>]*>`, report.Home.Body),
	)
	badgeClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<span[^>]*data-hard-case="badge"[^>]*>`, report.Home.Body),
	)
	mutedClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<span[^>]*data-hard-case="muted"[^>]*>`, report.Home.Body),
	)

	require.Equal(t, 200, report.RouteCSS.Status)
	require.Contains(t, report.RouteCSS.ContentType, "text/css")
	require.Contains(t, report.RouteCSS.Body, "padding:18px")
	require.Contains(t, report.RouteCSS.Body, "border:2px solid #2563eb")
	require.Contains(t, report.RouteCSS.Body, "outline:5px solid #f97316")
	require.Contains(t, report.RouteCSS.Body, "border-radius:9px")
	require.Contains(t, report.RouteCSS.Body, "background:#0f172a")
	require.Contains(t, report.RouteCSS.Body, "color:#047857")
	require.Equal(t, 1, strings.Count(report.RouteCSS.Body, "color:#047857"))
	requireNoOriginalClassSelector(t, report.RouteCSS.Body, "shell")
	requireNoOriginalClassSelector(t, report.RouteCSS.Body, "item")
	requireNoOriginalClassSelector(t, report.RouteCSS.Body, "featured")
	requireNoOriginalClassSelector(t, report.RouteCSS.Body, "label")
	requireNoOriginalClassSelector(t, report.RouteCSS.Body, "muted")
	requireNoOriginalClassSelector(t, report.RouteCSS.Body, "badge")
	requireNoOriginalClassSelector(t, report.RouteCSS.Body, "root")
	requireNoOriginalClassSelector(t, report.RouteCSS.Body, "value")

	shellClass := requireClassWithToken(t, report.RouteCSS.Body, shellClasses, "border:2px solid #2563eb")
	itemClass := requireClassWithToken(t, report.RouteCSS.Body, secondItemClasses, "padding-inline:7px")
	featuredClass := requireClassWithToken(t, report.RouteCSS.Body, secondItemClasses, "outline:5px solid #f97316")
	labelClass := requireClassWithToken(t, report.RouteCSS.Body, labelClasses, "text-transform:uppercase")
	badgeClass := requireClassWithToken(t, report.RouteCSS.Body, badgeClasses, "letter-spacing:.02em")
	mutedClass := requireClassWithToken(t, report.RouteCSS.Body, mutedClasses, "opacity:.66")
	require.Len(t, uniqueStrings([]string{shellClass, itemClass, featuredClass, labelClass, badgeClass, mutedClass}), 6)
	require.Regexp(
		t,
		regexp.MustCompile(
			`\.`+regexp.QuoteMeta(shellClass)+`>\.`+regexp.QuoteMeta(itemClass)+`\+\.`+
				regexp.QuoteMeta(itemClass)+`:is\(\.`+regexp.QuoteMeta(featuredClass)+
				`,\[data-state="\.featured"\]\) \.`+regexp.QuoteMeta(labelClass)+
				`:not\(\.`+regexp.QuoteMeta(mutedClass)+`\):{1,2}before\{[^}]*border-radius:9px`,
		),
		report.RouteCSS.Body,
	)
	require.Regexp(
		t,
		regexp.MustCompile(
			`@supports selector\(\.`+regexp.QuoteMeta(shellClass)+`:has\(\s*>\s*\.`+
				regexp.QuoteMeta(itemClass)+`\s*\+\s*\.`+regexp.QuoteMeta(itemClass)+
				`\.`+regexp.QuoteMeta(featuredClass)+`\)\)\{\.`+regexp.QuoteMeta(shellClass)+
				`:has\(\s*>\s*\.`+regexp.QuoteMeta(itemClass)+`\s*\+\s*\.`+
				regexp.QuoteMeta(itemClass)+`\.`+regexp.QuoteMeta(featuredClass)+
				`\)>\.`+regexp.QuoteMeta(badgeClass)+
				`:{1,2}after\{[^}]*background:#0f172a`,
		),
		report.RouteCSS.Body,
	)

	require.Equal(t, 200, report.RouteScript.Status)
	require.Contains(t, report.RouteScript.ContentType, "text/javascript")
	require.Contains(t, report.RouteScript.Body, "clientAssetsPage")
	require.Contains(t, report.RouteScript.Body, "clientAssetsMeter")
	require.Equal(t, 1, strings.Count(report.RouteScript.Body, "clientAssetsMeter"))
	require.NotContains(t, report.RouteScript.Body, "clientAssetsSectionLayout")
	require.NotContains(t, report.RouteScript.Body, "clientAssetsSectionPage")

	require.Equal(t, 200, report.About.Status)
	require.Contains(t, report.About.Body, `data-route-page="about"`)
	require.NotContains(t, report.About.Body, `/routes/index.css`)
	require.NotContains(t, report.About.Body, `/routes/index.js`)
	require.NotContains(t, report.About.Body, `/routes/section.css`)
	require.NotContains(t, report.About.Body, `/routes/section.js`)
	require.NotContains(t, report.About.Body, `/routes/complex.css`)

	require.Equal(t, 200, report.Section.Status)
	require.Contains(t, report.Section.Body, `data-layout="section"`)
	require.Contains(t, report.Section.Body, `data-route-page="section"`)
	require.Contains(t, report.Section.Body, report.SectionCSSURL)
	require.Contains(t, report.Section.Body, report.SectionScriptURL)
	require.Equal(t, 1, strings.Count(report.Section.Body, report.SectionCSSURL))
	require.Equal(t, 1, strings.Count(report.Section.Body, report.SectionScriptURL))
	require.NotContains(t, report.Section.Body, report.RouteCSSURL)
	require.NotContains(t, report.Section.Body, report.RouteScriptURL)
	require.NotContains(t, report.Section.Body, report.ComplexCSSURL)

	require.Equal(t, 200, report.SectionCSS.Status)
	require.Contains(t, report.SectionCSS.ContentType, "text/css")
	require.Contains(t, report.SectionCSS.Body, "background:#fef3c7")
	require.Contains(t, report.SectionCSS.Body, "outline:4px solid #7c3aed")
	require.Contains(t, report.SectionCSS.Body, "color:#047857")
	require.Equal(t, 1, strings.Count(report.SectionCSS.Body, "color:#047857"))
	require.NotContains(t, report.SectionCSS.Body, ".shell")
	require.NotContains(t, report.SectionCSS.Body, ".panel")
	require.NotContains(t, report.SectionCSS.Body, ".root")
	require.NotContains(t, report.SectionCSS.Body, ".value")

	require.Equal(t, 200, report.SectionScript.Status)
	require.Contains(t, report.SectionScript.ContentType, "text/javascript")
	require.Contains(t, report.SectionScript.Body, "clientAssetsSectionLayout")
	require.Contains(t, report.SectionScript.Body, "clientAssetsSectionPage")
	require.Contains(t, report.SectionScript.Body, "clientAssetsMeter")
	require.Equal(t, 1, strings.Count(report.SectionScript.Body, "clientAssetsMeter"))
	require.NotContains(t, report.SectionScript.Body, "clientAssetsPage")

	requireComplexClientAssetCSS(t, report)

	require.Equal(t, 404, report.NotFound.Status)
	require.Contains(t, report.NotFound.Body, report.NotFoundCSSURL)
	require.Equal(t, 1, strings.Count(report.NotFound.Body, report.NotFoundCSSURL))
	require.NotContains(t, report.NotFound.Body, report.RouteScriptURL)
	require.NotContains(t, report.NotFound.Body, report.SectionScriptURL)
	require.Equal(t, 200, report.NotFoundCSS.Status)
	require.Contains(t, report.NotFoundCSS.Body, "border:3px dashed #b42318")

	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "page.css_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "page.ts_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "section", "layout.css_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "section", "layout.ts_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "section", "page.css_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "section", "page.ts_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "complex", "page.css_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "components", "meter", "meter.css_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "components", "meter", "meter.ts_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "components", "filterpanel", "filter_panel.css_gen.go"))
}

func requireComplexClientAssetCSS(t *testing.T, report clientAssetsFixture) {
	t.Helper()

	require.Equal(t, 200, report.Complex.Status)
	require.Contains(t, report.Complex.Body, `data-route-page="complex"`)
	require.Contains(t, report.Complex.Body, report.ComplexCSSURL)
	require.Equal(t, 1, strings.Count(report.Complex.Body, report.ComplexCSSURL))
	require.NotContains(t, report.Complex.Body, report.RouteCSSURL)
	require.NotContains(t, report.Complex.Body, report.SectionCSSURL)

	dashboardClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<main[^>]*data-route-page="complex"[^>]*>`, report.Complex.Body),
	)
	railClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<nav[^>]*data-complex="rail"[^>]*>`, report.Complex.Body),
	)
	cardClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<article[^>]*data-complex="card"[^>]*>`, report.Complex.Body),
	)
	statusClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<span[^>]*data-complex="status"[^>]*>`, report.Complex.Body),
	)
	metaClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<span[^>]*data-complex="meta"[^>]*>`, report.Complex.Body),
	)
	linkClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<a[^>]*data-complex="link"[^>]*>`, report.Complex.Body),
	)
	filterPanelClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<aside[^>]*data-component="filter-panel"[^>]*>`, report.Complex.Body),
	)
	filterClusterClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<ul[^>]*data-filter-cluster="primary"[^>]*>`, report.Complex.Body),
	)
	filterRiskClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<li[^>]*data-filter-token="risk"[^>]*>`, report.Complex.Body),
	)

	require.Equal(t, 200, report.ComplexCSS.Status)
	require.Contains(t, report.ComplexCSS.ContentType, "text/css")
	require.Contains(t, report.ComplexCSS.Body, "border:4px solid #0284c7")
	require.Contains(t, report.ComplexCSS.Body, "box-shadow:0 0 0 3px #0ea5e9")
	require.Contains(t, report.ComplexCSS.Body, "border-block-end:5px solid #db2777")
	require.Contains(t, report.ComplexCSS.Body, "background:#f59e0b")
	require.Contains(t, report.ComplexCSS.Body, "outline:6px solid #14b8a6")
	require.Contains(t, report.ComplexCSS.Body, "text-shadow:0 1px #111827")
	require.Contains(t, report.ComplexCSS.Body, "@layer clientassets.e2e")
	require.NotContains(t, report.ComplexCSS.Body, "@layer clientassets.n_")
	require.Contains(t, report.ComplexCSS.Body, "grid-template-columns:repeat(2,minmax(0,1fr))")
	require.Contains(t, report.ComplexCSS.Body, "box-shadow:0 0 0 2px #16a34a")

	for _, className := range []string{
		"dashboard",
		"rail",
		"card",
		"primary",
		"secondary",
		"disabled",
		"status",
		"meta",
		"link",
		"ghost",
		"panel",
		"cluster",
		"token",
		"active",
		"danger",
		"expanded",
		"collapsed",
	} {
		requireNoOriginalClassSelector(t, report.ComplexCSS.Body, className)
	}

	dashboardClass := requireClassWithToken(t, report.ComplexCSS.Body, dashboardClasses, "color:#111827")
	railClass := requireClassWithToken(t, report.ComplexCSS.Body, railClasses, "order:1")
	cardClass := requireClassWithToken(t, report.ComplexCSS.Body, cardClasses, "border:4px solid #0284c7")
	primaryClass := requireClassWithToken(t, report.ComplexCSS.Body, cardClasses, "background:#e0f2fe")
	statusClass := requireClassWithToken(t, report.ComplexCSS.Body, statusClasses, "font-weight:700")
	metaClass := requireClassWithToken(t, report.ComplexCSS.Body, metaClasses, "color:#7c2d12")
	linkClass := requireClassWithToken(t, report.ComplexCSS.Body, linkClasses, "text-decoration-thickness:3px")
	filterPanelClass := requireClassWithToken(
		t,
		report.ComplexCSS.Body,
		filterPanelClasses,
		"grid-template-columns:repeat(2,minmax(0,1fr))",
	)
	filterClusterClass := requireClassWithToken(t, report.ComplexCSS.Body, filterClusterClasses, "display:contents")
	filterExpandedClass := requireClassWithToken(
		t,
		report.ComplexCSS.Body,
		filterClusterClasses,
		"border-inline-start:3px solid #22c55e",
	)
	filterTokenClass := requireClassWithToken(t, report.ComplexCSS.Body, filterRiskClasses, "padding:4px 8px")
	filterDangerClass := requireClassWithToken(t, report.ComplexCSS.Body, filterRiskClasses, "background:#fee2e2")

	require.Len(t, uniqueStrings([]string{
		dashboardClass,
		railClass,
		cardClass,
		primaryClass,
		statusClass,
		metaClass,
		linkClass,
		filterPanelClass,
		filterClusterClass,
		filterExpandedClass,
		filterTokenClass,
		filterDangerClass,
	}), 12)

	require.Regexp(
		t,
		regexp.MustCompile(
			`\.`+regexp.QuoteMeta(dashboardClass)+`\[data-theme="\.ghost"\]>\.`+
				regexp.QuoteMeta(railClass)+`\+\.`+regexp.QuoteMeta(cardClass)+
				`:is\(\.`+regexp.QuoteMeta(primaryClass)+`,\.[^)]+\):not\(\.[^)]+\) \.`+
				regexp.QuoteMeta(statusClass)+`:{1,2}before\{[^}]*box-shadow:0 0 0 3px #0ea5e9`,
		),
		report.ComplexCSS.Body,
	)
	require.Regexp(
		t,
		regexp.MustCompile(
			`@media\s*\(min-width:42rem\)\{\.`+regexp.QuoteMeta(dashboardClass)+`>\.`+
				regexp.QuoteMeta(railClass)+`~\.`+regexp.QuoteMeta(cardClass)+
				`:has\(>\.`+regexp.QuoteMeta(statusClass)+`\+\.`+regexp.QuoteMeta(metaClass)+
				`\[data-state~="\.hot"\]\)>\.`+regexp.QuoteMeta(linkClass)+
				`:any-link:{1,2}after\{[^}]*border-block-end:5px solid #db2777`,
		),
		report.ComplexCSS.Body,
	)
	require.Regexp(
		t,
		regexp.MustCompile(
			`@container \(min-width:45rem\)\{\.`+regexp.QuoteMeta(dashboardClass)+
				`:has\(>\.`+regexp.QuoteMeta(cardClass)+`\.`+regexp.QuoteMeta(primaryClass)+
				`\)>\.`+regexp.QuoteMeta(cardClass)+` \.`+regexp.QuoteMeta(metaClass)+
				`:{1,2}after\{[^}]*background:#f59e0b`,
		),
		report.ComplexCSS.Body,
	)
	require.Regexp(
		t,
		regexp.MustCompile(
			`@scope\s*\(\.`+regexp.QuoteMeta(dashboardClass)+`\)\s*to\s*\(\.[^)]+\)\{\.`+
				regexp.QuoteMeta(cardClass)+`>\.`+regexp.QuoteMeta(statusClass)+
				`\{[^}]*text-shadow:0 1px #111827`,
		),
		report.ComplexCSS.Body,
	)
	require.Regexp(
		t,
		regexp.MustCompile(
			`@container filter-panel \(min-width:20rem\)\{\.`+regexp.QuoteMeta(filterPanelClass)+
				`:has\(\.`+regexp.QuoteMeta(filterTokenClass)+`\.[^)]+\) \.`+
				regexp.QuoteMeta(filterClusterClass)+`>\.`+regexp.QuoteMeta(filterTokenClass)+
				`\+\.`+regexp.QuoteMeta(filterTokenClass)+`\{[^}]*box-shadow:0 0 0 2px #16a34a`,
		),
		report.ComplexCSS.Body,
	)
}

func TestTemplCSSFixtureApp(t *testing.T) {
	report := loadTemplCSSFixture(t)

	require.Regexp(
		t,
		regexp.MustCompile(`^/_assets/[0-9a-f]{16}/styles/templ\.css$`),
		report.StylesheetURL,
	)

	require.Equal(t, 200, report.Home.Status)
	require.Contains(t, report.Home.ContentType, "text/html")
	require.Contains(t, report.Home.Body, `<html lang="">`)
	require.Contains(t, report.Home.Body, `id="root-layout"`)
	require.Contains(t, report.Home.Body, `E2E Home`)
	require.Contains(t, report.Home.Body, `Shared component body`)
	require.Contains(t, report.Home.Body, report.StylesheetURL)
	require.NotContains(t, report.Home.Body, `<style type="text/css"`)

	require.Equal(t, 200, report.Partial.Status)
	require.Contains(t, report.Partial.ContentType, "text/html")
	require.NotContains(t, report.Partial.Body, `<html`)
	require.Contains(t, report.Partial.Body, `E2E Home`)
	require.Contains(t, report.Partial.Body, `Shared component body`)
	require.Contains(t, report.Partial.HXTriggerAfterSettle, `styles/templ.css`)
	require.NotContains(t, report.Partial.Body, `<style type="text/css"`)

	require.Equal(t, 404, report.NotFound.Status)
	require.Contains(t, report.NotFound.Body, `id="not-found"`)
	require.Contains(t, report.NotFound.Body, `Missing /missing`)
	require.Contains(t, report.NotFound.Body, report.StylesheetURL)

	require.Equal(t, 200, report.Stylesheet.Status)
	require.Contains(t, report.Stylesheet.ContentType, "text/css")
	require.Contains(t, report.Stylesheet.Body, `border:1px solid #123`)
	require.Contains(t, report.Stylesheet.Body, `width:50%`)
	require.Contains(t, report.Stylesheet.Body, `padding:12px`)
}

func TestNamespacedRoutesTemplCSSFixtureApp(t *testing.T) {
	report := loadNamespacedFixture(t)

	require.Regexp(
		t,
		regexp.MustCompile(`^/_assets/[0-9a-f]{16}/styles/templ\.css$`),
		report.StylesheetURL,
	)

	require.Equal(t, 200, report.Dashboard.Status)
	require.Contains(t, report.Dashboard.ContentType, "text/html")
	require.Contains(t, report.Dashboard.Body, `<html lang="">`)
	require.Contains(t, report.Dashboard.Body, `id="root-layout"`)
	require.Contains(t, report.Dashboard.Body, `id="marketing-dashboard"`)
	require.Contains(t, report.Dashboard.Body, `Marketing Dashboard`)
	require.Contains(t, report.Dashboard.Body, `data-slot="analytics"`)
	require.Contains(t, report.Dashboard.Body, `Quarterly analytics`)
	require.Contains(t, report.Dashboard.Body, `data-component="stat-chip"`)
	require.Contains(t, report.Dashboard.Body, report.StylesheetURL)
	require.NotContains(t, report.Dashboard.Body, `<style type="text/css"`)

	require.Equal(t, 200, report.Partial.Status)
	require.Contains(t, report.Partial.ContentType, "text/html")
	require.NotContains(t, report.Partial.Body, `<html`)
	require.Contains(t, report.Partial.Body, `id="marketing-dashboard"`)
	require.NotContains(t, report.Partial.Body, `data-slot="analytics"`)
	require.Contains(t, report.Partial.HXTriggerAfterSettle, `styles/templ.css`)
	require.NotContains(t, report.Partial.Body, `<style type="text/css"`)

	require.Equal(t, 404, report.NotFound.Status)
	require.Contains(t, report.NotFound.Body, `id="not-found"`)
	require.Contains(t, report.NotFound.Body, `Missing /unknown`)
	require.Contains(t, report.NotFound.Body, report.StylesheetURL)

	require.Equal(t, 200, report.Stylesheet.Status)
	require.Contains(t, report.Stylesheet.ContentType, "text/css")
	require.Contains(t, report.Stylesheet.Body, `border:2px solid #17324d`)
	require.Contains(t, report.Stylesheet.Body, `letter-spacing:.08em`)
	require.Contains(t, report.Stylesheet.Body, `width:72%`)
	require.Contains(t, report.Stylesheet.Body, `padding:14px`)
}

func TestDocsFeatureFixtureApp(t *testing.T) {
	report := loadDocsFeatureFixture(t)

	require.Regexp(
		t,
		regexp.MustCompile(`^/_assets/[0-9a-f]{16}/site\.css$`),
		report.SiteCSSURL,
	)
	require.Regexp(
		t,
		regexp.MustCompile(`^/_assets/[0-9a-f]{16}/styles/templ\.css$`),
		report.TemplCSSURL,
	)

	require.Equal(t, 200, report.AuthorDE.Status)
	require.Contains(t, report.AuthorDE.ContentType, "text/html")
	require.Contains(t, report.AuthorDE.Body, `<html lang="de">`)
	require.Contains(t, report.AuthorDE.Body, `id="author-page"`)
	require.Contains(t, report.AuthorDE.Body, `Autor ada`)
	require.Contains(t, report.AuthorDE.Body, `Notizen von ada.`)
	require.Contains(t, report.AuthorDE.Body, `cached:ada:1|cached:ada:1`)
	require.Contains(
		t,
		report.AuthorDE.Body,
		`id="switch-en" href="https://docs.example.test/blog/author/ada"`,
	)
	require.Contains(
		t,
		report.AuthorDE.Body,
		`rel="canonical" href="https://docs.example.test/blog/de/author/ada"`,
	)
	require.Contains(
		t,
		report.AuthorDE.Body,
		`hreflang="en" href="https://docs.example.test/blog/author/ada"`,
	)
	require.Contains(
		t,
		report.AuthorDE.Body,
		`hreflang="de" href="https://docs.example.test/blog/de/author/ada"`,
	)
	require.Contains(
		t,
		report.AuthorDE.Body,
		`property="og:url" content="https://docs.example.test/blog/de/author/ada"`,
	)
	require.Contains(t, report.AuthorDE.Body, report.SiteCSSURL)
	require.Contains(t, report.AuthorDE.Body, report.TemplCSSURL)
	require.NotContains(t, report.AuthorDE.Body, `<style type="text/css"`)

	require.Equal(t, 1, report.ExpensiveLoadsAfterFirst)

	require.Equal(t, 308, report.AuthorENRedirect.Status)
	require.Equal(t, "/author/ada", report.AuthorENRedirect.Location)

	require.Equal(t, 200, report.AuthorEN.Status)
	require.Contains(t, report.AuthorEN.Body, `<html lang="en">`)
	require.Contains(t, report.AuthorEN.Body, `Author ada`)
	require.Contains(t, report.AuthorEN.Body, `Browse notes by ada.`)
	require.Equal(t, 2, report.ExpensiveLoadsAfterSecond)

	require.Equal(t, 404, report.AuthorMissing.Status)
	require.Contains(t, report.AuthorMissing.Body, `id="author-not-found"`)
	require.Contains(t, report.AuthorMissing.Body, `Unknown author /author/missing`)

	require.Equal(t, 500, report.AuthorError.Status)
	require.Contains(t, report.AuthorError.ContentType, "text/plain")
	require.Contains(t, report.AuthorError.Body, `Internal Server Error`)

	require.Equal(t, 200, report.Dashboard.Status)
	require.Contains(t, report.Dashboard.Body, `<html lang="de">`)
	require.Contains(t, report.Dashboard.Body, `id="marketing-dashboard"`)
	require.Contains(t, report.Dashboard.Body, `Quarterly analytics`)
	require.Contains(t, report.Dashboard.Body, `data-component="profile-card"`)

	require.Equal(t, 200, report.DashboardPartial.Status)
	require.NotContains(t, report.DashboardPartial.Body, `<html`)
	require.Contains(t, report.DashboardPartial.Body, `id="marketing-dashboard"`)
	require.NotContains(t, report.DashboardPartial.Body, `Quarterly analytics`)
	require.Contains(t, report.DashboardPartial.HXTriggerAfterSettle, `styles/templ.css`)

	require.Equal(t, 200, report.PingDE.Status)
	require.Contains(t, report.PingDE.ContentType, "application/json")
	require.Contains(t, report.PingDE.Body, `"locale":"de"`)
	require.Contains(t, report.PingDE.Body, `"path":"/api/ping"`)

	require.Equal(t, 200, report.Robots.Status)
	require.Contains(t, report.Robots.ContentType, "text/plain")
	require.Contains(t, report.Robots.Body, "User-agent: *")
	require.Contains(t, report.Robots.Body, "Disallow: /api")
	require.Contains(t, report.Robots.Body, "Host: docs.example.test")
	require.Contains(t, report.Robots.Body, "Sitemap: https://docs.example.test/blog/sitemap-index.xml")

	require.Equal(t, 200, report.Feed.Status)
	require.Contains(t, report.Feed.ContentType, "application/rss+xml")
	require.Contains(t, report.Feed.Body, "<title>Docs Feed</title>")
	require.Contains(t, report.Feed.Body, "https://docs.example.test/blog/feed.xml")

	require.Equal(t, 200, report.AuthorFeed.Status)
	require.Contains(t, report.AuthorFeed.ContentType, "application/rss+xml")
	require.Contains(t, report.AuthorFeed.Body, "<title>Feed for ada</title>")
	require.Contains(t, report.AuthorFeed.Body, "https://docs.example.test/blog/author/ada")

	require.Equal(t, 200, report.Sitemap.Status)
	require.Contains(t, report.Sitemap.ContentType, "application/xml")
	require.Contains(t, report.Sitemap.Body, "https://docs.example.test/blog/dashboard")

	require.Equal(t, 200, report.SitemapIndex.Status)
	require.Contains(t, report.SitemapIndex.ContentType, "application/xml")
	require.Contains(t, report.SitemapIndex.Body, "https://docs.example.test/blog/sitemap/authors.xml")

	require.Equal(t, 200, report.SitemapChunk.Status)
	require.Contains(t, report.SitemapChunk.ContentType, "application/xml")
	require.Contains(t, report.SitemapChunk.Body, "https://docs.example.test/blog/author/ada")
	require.Contains(t, report.SitemapChunk.Body, "https://docs.example.test/blog/de/author/ada")

	require.Equal(t, 200, report.Favicon.Status)
	require.Contains(t, report.Favicon.Body, "fixture-icon")

	require.Equal(t, 200, report.SiteCSS.Status)
	require.Contains(t, report.SiteCSS.ContentType, "text/css")
	require.Contains(t, report.SiteCSS.Body, "font-family:monospace")
	require.Contains(t, report.SiteCSS.Body, "background:#f5f7fb")

	require.Equal(t, 200, report.TemplCSS.Status)
	require.Contains(t, report.TemplCSS.ContentType, "text/css")
	require.Contains(t, report.TemplCSS.Body, "border:1px solid #224")
	require.Contains(t, report.TemplCSS.Body, "width:72%")
	require.Contains(t, report.TemplCSS.Body, "width:64%")
	require.Contains(t, report.TemplCSS.Body, "letter-spacing:.08em")

	require.Equal(t, 200, report.Health.Status)
	require.Equal(t, "ok", strings.TrimSpace(report.Health.Body))
}

func TestGroupedNamespaceFixtureApp(t *testing.T) {
	report := loadGroupedNamespaceFixture(t)

	require.Regexp(
		t,
		regexp.MustCompile(`^/_assets/[0-9a-f]{16}/styles/templ\.css$`),
		report.StylesheetURL,
	)

	t.Run("notes", func(t *testing.T) {
		require.Equal(t, 200, report.Notes.Status)
		require.Contains(t, report.Notes.ContentType, "text/html")
		require.Contains(t, report.Notes.Body, `<html lang="">`)
		require.Contains(t, report.Notes.Body, `data-layout="blog-discover"`)
		require.NotContains(t, report.Notes.Body, `data-layout="shop-discover"`)
		require.NotContains(t, report.Notes.Body, `data-layout="editorial-discover"`)
		require.Contains(t, report.Notes.Body, `id="discover-notes"`)
		require.Contains(t, report.Notes.Body, `>Notes<`)
		require.Contains(t, report.Notes.Body, report.StylesheetURL)
		require.NotContains(t, report.Notes.Body, `<style type="text/css"`)

		blogClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<section[^>]*data-layout="blog-discover"[^>]*>`, report.Notes.Body),
		)
		requireClassWithToken(t, report.Stylesheet.Body, blogClasses, "border:2px solid #17324d")

		notesClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<main[^>]*id="discover-notes"[^>]*>`, report.Notes.Body),
		)
		requireClassWithToken(t, report.Stylesheet.Body, notesClasses, "background:#eef5ff")

		progressClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<div[^>]*data-progress="64"[^>]*>`, report.Notes.Body),
		)
		requireClassWithToken(t, report.Stylesheet.Body, progressClasses, "width:64%")
	})

	t.Run("notes_partial", func(t *testing.T) {
		require.Equal(t, 200, report.NotesPartial.Status)
		require.NotContains(t, report.NotesPartial.Body, `<html`)
		require.Contains(t, report.NotesPartial.Body, `id="discover-notes"`)
		require.NotContains(t, report.NotesPartial.Body, `data-layout="blog-discover"`)
		require.Contains(t, report.NotesPartial.HXTriggerAfterSettle, `styles/templ.css`)
	})

	t.Run("guides", func(t *testing.T) {
		require.Equal(t, 200, report.Guides.Status)
		require.Contains(t, report.Guides.Body, `data-layout="blog-discover"`)
		require.Contains(t, report.Guides.Body, `data-layout="editorial-discover"`)
		require.NotContains(t, report.Guides.Body, `data-layout="shop-discover"`)
		require.Contains(t, report.Guides.Body, `id="discover-guides"`)
		require.NotContains(t, report.Guides.Body, `<style type="text/css"`)

		editorialClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<section[^>]*data-layout="editorial-discover"[^>]*>`, report.Guides.Body),
		)
		requireClassWithToken(t, report.Stylesheet.Body, editorialClasses, "background:#effcf6")

		guidesClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<main[^>]*id="discover-guides"[^>]*>`, report.Guides.Body),
		)
		requireClassWithToken(t, report.Stylesheet.Body, guidesClasses, "border:1px dashed #17594a")
	})

	t.Run("tags", func(t *testing.T) {
		require.Equal(t, 200, report.Tags.Status)
		require.Contains(t, report.Tags.ContentType, "text/html")
		require.Contains(t, report.Tags.Body, `data-layout="shop-discover"`)
		require.NotContains(t, report.Tags.Body, `data-layout="blog-discover"`)
		require.NotContains(t, report.Tags.Body, `data-layout="editorial-discover"`)
		require.Contains(t, report.Tags.Body, `id="discover-tags"`)
		require.Contains(t, report.Tags.Body, `>Tags<`)
		require.NotContains(t, report.Tags.Body, `<style type="text/css"`)

		shopClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<section[^>]*data-layout="shop-discover"[^>]*>`, report.Tags.Body),
		)
		requireClassWithToken(t, report.Stylesheet.Body, shopClasses, "background:#fff7ed")

		tagsClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<main[^>]*id="discover-tags"[^>]*>`, report.Tags.Body),
		)
		requireClassWithToken(t, report.Stylesheet.Body, tagsClasses, "background:#fff1e6")

		progressClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<div[^>]*data-progress="80"[^>]*>`, report.Tags.Body),
		)
		requireClassWithToken(t, report.Stylesheet.Body, progressClasses, "width:80%")
	})

	t.Run("tags_partial", func(t *testing.T) {
		require.Equal(t, 200, report.TagsPartial.Status)
		require.NotContains(t, report.TagsPartial.Body, `<html`)
		require.Contains(t, report.TagsPartial.Body, `id="discover-tags"`)
		require.NotContains(t, report.TagsPartial.Body, `data-layout="shop-discover"`)
		require.Contains(t, report.TagsPartial.HXTriggerAfterSettle, `styles/templ.css`)
	})

	t.Run("stylesheet", func(t *testing.T) {
		require.Equal(t, 200, report.Stylesheet.Status)
		require.Contains(t, report.Stylesheet.ContentType, "text/css")
		require.Contains(t, report.Stylesheet.Body, "border:2px solid #17324d")
		require.Contains(t, report.Stylesheet.Body, "background:#fff7ed")
		require.Contains(t, report.Stylesheet.Body, "width:64%")
		require.Contains(t, report.Stylesheet.Body, "width:80%")
		require.Contains(t, report.Stylesheet.Body, "letter-spacing:.08em")
	})
}

func TestCatchAllFixtureApp(t *testing.T) {
	report := loadCatchAllFixture(t)
	require.Regexp(
		t,
		regexp.MustCompile(`^/_assets/[0-9a-f]{16}/styles/templ\.css$`),
		report.StylesheetURL,
	)

	require.Equal(t, 200, report.Docs.Status)
	require.Contains(t, report.Docs.ContentType, "text/html")
	require.Contains(t, report.Docs.Body, `<html lang="">`)
	require.Contains(t, report.Docs.Body, `id="docs-catchall"`)
	require.Contains(t, report.Docs.Body, `data-joined>a/b<`)
	require.Contains(t, report.Docs.Body, `data-depth>2<`)
	require.Contains(t, report.Docs.Body, report.StylesheetURL)
	require.NotContains(t, report.Docs.Body, `<style type="text/css"`)

	docsClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<main[^>]*id="docs-catchall"[^>]*>`, report.Docs.Body),
	)
	requireClassWithToken(t, report.Stylesheet.Body, docsClasses, "background:#eef5ff")

	require.Equal(t, 200, report.Nested.Status)
	require.Contains(t, report.Nested.Body, `data-joined>alpha/beta/gamma<`)
	require.Contains(t, report.Nested.Body, `data-depth>3<`)

	require.Equal(t, 200, report.Partial.Status)
	require.NotContains(t, report.Partial.Body, `<html`)
	require.Contains(t, report.Partial.Body, `id="docs-catchall"`)
	require.Contains(t, report.Partial.HXTriggerAfterSettle, `styles/templ.css`)

	require.Equal(t, 404, report.Missing.Status)
	require.Contains(t, report.Missing.Body, `id="not-found"`)
	require.Contains(t, report.Missing.Body, `Missing /docs`)

	require.Equal(t, 200, report.Stylesheet.Status)
	require.Contains(t, report.Stylesheet.ContentType, "text/css")
	require.Contains(t, report.Stylesheet.Body, "background:#eef5ff")
	require.Contains(t, report.Stylesheet.Body, "padding:12px")
}

func TestOptionalCatchAllFixtureApp(t *testing.T) {
	report := loadOptionalCatchAllFixture(t)
	require.Regexp(
		t,
		regexp.MustCompile(`^/_assets/[0-9a-f]{16}/styles/templ\.css$`),
		report.StylesheetURL,
	)

	require.Equal(t, 200, report.Root.Status)
	require.Contains(t, report.Root.ContentType, "text/html")
	require.Contains(t, report.Root.Body, `<html lang="">`)
	require.Contains(t, report.Root.Body, `id="library-page"`)
	require.Contains(t, report.Root.Body, `data-joined>root<`)
	require.Contains(t, report.Root.Body, `data-depth>0<`)
	require.Contains(t, report.Root.Body, report.StylesheetURL)
	require.NotContains(t, report.Root.Body, `<style type="text/css"`)

	rootClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<main[^>]*id="library-page"[^>]*>`, report.Root.Body),
	)
	requireClassWithToken(t, report.Stylesheet.Body, rootClasses, "background:#fff7ed")

	require.Equal(t, 200, report.Nested.Status)
	require.Contains(t, report.Nested.Body, `data-joined>a/b<`)
	require.Contains(t, report.Nested.Body, `data-depth>2<`)

	require.Equal(t, 200, report.Partial.Status)
	require.NotContains(t, report.Partial.Body, `<html`)
	require.Contains(t, report.Partial.Body, `id="library-page"`)
	require.Contains(t, report.Partial.HXTriggerAfterSettle, `styles/templ.css`)

	require.Equal(t, 200, report.Stylesheet.Status)
	require.Contains(t, report.Stylesheet.ContentType, "text/css")
	require.Contains(t, report.Stylesheet.Body, "background:#fff7ed")
	require.Contains(t, report.Stylesheet.Body, "padding:12px")
}

func TestMethodMatrixFixtureApp(t *testing.T) {
	report := loadMethodMatrixFixture(t)
	require.Regexp(
		t,
		regexp.MustCompile(`^/_assets/[0-9a-f]{16}/styles/templ\.css$`),
		report.StylesheetURL,
	)

	require.Equal(t, 200, report.Home.Status)
	require.Contains(t, report.Home.ContentType, "text/html")
	require.Contains(t, report.Home.Body, `<html lang="">`)
	require.Contains(t, report.Home.Body, `id="method-home"`)
	require.Contains(t, report.Home.Body, report.StylesheetURL)
	require.NotContains(t, report.Home.Body, `<style type="text/css"`)

	summaryClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<section[^>]*data-component="summary"[^>]*>`, report.Home.Body),
	)
	requireClassWithToken(t, report.Stylesheet.Body, summaryClasses, "border:1px solid #17324d")

	require.Equal(t, 200, report.Get.Status)
	require.Contains(t, report.Get.ContentType, "application/json")
	require.Contains(t, report.Get.Body, `"method":"GET"`)
	require.Contains(t, report.Get.Body, `"slug":"ada"`)
	require.Contains(t, report.Get.Body, `"path":"/api/note/ada"`)

	require.Equal(t, 200, report.Head.Status)
	require.Contains(t, report.Head.ContentType, "application/json")
	require.Empty(t, strings.TrimSpace(report.Head.Body))

	require.Equal(t, 204, report.Options.Status)
	require.Equal(t, "GET, POST, PATCH, DELETE, HEAD, OPTIONS", report.Options.Allow)
	require.Empty(t, strings.TrimSpace(report.Options.Body))

	require.Equal(t, 201, report.Post.Status)
	require.Equal(t, "/api/note/ada", report.Post.Location)
	require.Contains(t, report.Post.Body, `"method":"POST"`)

	require.Equal(t, 202, report.Patch.Status)
	require.Contains(t, report.Patch.Body, `"method":"PATCH"`)

	require.Equal(t, 204, report.Delete.Status)
	require.Empty(t, strings.TrimSpace(report.Delete.Body))

	require.Equal(t, 405, report.Put.Status)
	require.Equal(t, "GET, POST, PATCH, DELETE, HEAD, OPTIONS", report.Put.Allow)
	require.Contains(t, report.Put.Body, "Method Not Allowed")

	require.Equal(t, 404, report.Missing.Status)
	require.Contains(t, report.Missing.Body, `id="not-found"`)
	require.Contains(t, report.Missing.Body, `Missing /api/missing`)

	require.Equal(t, 200, report.Stylesheet.Status)
	require.Contains(t, report.Stylesheet.ContentType, "text/css")
	require.Contains(t, report.Stylesheet.Body, "background:#eef5ff")
	require.Contains(t, report.Stylesheet.Body, "border:1px solid #17324d")
}

func TestI18nPrefixAlwaysFixtureApp(t *testing.T) {
	report := loadPrefixAlwaysFixture(t)
	require.Regexp(
		t,
		regexp.MustCompile(`^/_assets/[0-9a-f]{16}/styles/templ\.css$`),
		report.StylesheetURL,
	)

	require.Equal(t, 308, report.RootRedirect.Status)
	require.Equal(t, "/en", report.RootRedirect.Location)

	require.Equal(t, 200, report.HomeEN.Status)
	require.Contains(t, report.HomeEN.ContentType, "text/html")
	require.Contains(t, report.HomeEN.Body, `<html lang="en">`)
	require.Contains(t, report.HomeEN.Body, `Localized Home`)
	require.Contains(t, report.HomeEN.Body, `data-locale="en"`)
	require.Contains(t, report.HomeEN.Body, `id="switch-de" href="https://prefix.example.test/de"`)
	require.Contains(t, report.HomeEN.Body, report.StylesheetURL)

	require.Equal(t, 200, report.HomeDE.Status)
	require.Contains(t, report.HomeDE.Body, `<html lang="de">`)
	require.Contains(t, report.HomeDE.Body, `Lokales Zuhause`)
	require.Contains(t, report.HomeDE.Body, `data-locale="de"`)
	require.Contains(t, report.HomeDE.Body, `id="switch-en" href="https://prefix.example.test/en"`)

	require.Equal(t, 404, report.NotFoundEN.Status)
	require.Contains(t, report.NotFoundEN.Body, `<html lang="en">`)
	require.Contains(t, report.NotFoundEN.Body, `Page not found /en/missing`)
	require.Contains(t, report.NotFoundEN.Body, `data-locale="en"`)
	require.Contains(t, report.NotFoundEN.Body, `data-request-path="/en/missing"`)
	require.Contains(t, report.NotFoundEN.Body, `data-not-found-source="unmatched_route"`)

	require.Equal(t, 404, report.NotFoundDE.Status)
	require.Contains(t, report.NotFoundDE.Body, `<html lang="de">`)
	require.Contains(t, report.NotFoundDE.Body, `Seite nicht gefunden /de/missing`)
	require.Contains(t, report.NotFoundDE.Body, `data-locale="de"`)
	require.Contains(t, report.NotFoundDE.Body, `data-request-path="/de/missing"`)
	require.Contains(t, report.NotFoundDE.Body, `data-not-found-source="unmatched_route"`)

	require.Equal(t, 404, report.PageLoadNotFoundEN.Status)
	require.Contains(t, report.PageLoadNotFoundEN.Body, `<html lang="en">`)
	require.Contains(t, report.PageLoadNotFoundEN.Body, `Page not found /fail`)
	require.Contains(t, report.PageLoadNotFoundEN.Body, `data-locale="en"`)
	require.Contains(t, report.PageLoadNotFoundEN.Body, `data-request-path="/fail"`)
	require.Contains(t, report.PageLoadNotFoundEN.Body, `data-not-found-source="page_load"`)

	require.Equal(t, 404, report.PageLoadNotFoundDE.Status)
	require.Contains(t, report.PageLoadNotFoundDE.Body, `<html lang="de">`)
	require.Contains(t, report.PageLoadNotFoundDE.Body, `Seite nicht gefunden /fail`)
	require.Contains(t, report.PageLoadNotFoundDE.Body, `data-locale="de"`)
	require.Contains(t, report.PageLoadNotFoundDE.Body, `data-request-path="/fail"`)
	require.Contains(t, report.PageLoadNotFoundDE.Body, `data-not-found-source="page_load"`)

	require.Equal(t, 404, report.HelpNotFoundEN.Status)
	require.Contains(t, report.HelpNotFoundEN.Body, `<html lang="en">`)
	require.Contains(t, report.HelpNotFoundEN.Body, `data-layout="support-help"`)
	require.Contains(t, report.HelpNotFoundEN.Body, `data-layout-locale="en"`)
	require.Contains(t, report.HelpNotFoundEN.Body, `id="help-not-found"`)
	require.Contains(t, report.HelpNotFoundEN.Body, `Support help: Page not found /help/fail`)
	require.Contains(t, report.HelpNotFoundEN.Body, `data-locale="en"`)
	require.Contains(t, report.HelpNotFoundEN.Body, `data-request-path="/help/fail"`)
	require.Contains(t, report.HelpNotFoundEN.Body, `data-not-found-source="page_load"`)

	require.Equal(t, 404, report.HelpNotFoundDE.Status)
	require.Contains(t, report.HelpNotFoundDE.Body, `<html lang="de">`)
	require.Contains(t, report.HelpNotFoundDE.Body, `data-layout="support-help"`)
	require.Contains(t, report.HelpNotFoundDE.Body, `data-layout-locale="de"`)
	require.Contains(t, report.HelpNotFoundDE.Body, `id="help-not-found"`)
	require.Contains(t, report.HelpNotFoundDE.Body, `Support help: Seite nicht gefunden /help/fail`)
	require.Contains(t, report.HelpNotFoundDE.Body, `data-locale="de"`)
	require.Contains(t, report.HelpNotFoundDE.Body, `data-request-path="/help/fail"`)
	require.Contains(t, report.HelpNotFoundDE.Body, `data-not-found-source="page_load"`)

	require.Equal(t, 308, report.GreetRedirect.Status)
	require.Equal(t, "/en/greet/ada", report.GreetRedirect.Location)

	require.Equal(t, 200, report.GreetEN.Status)
	require.Contains(t, report.GreetEN.Body, `<html lang="en">`)
	require.Contains(t, report.GreetEN.Body, `id="greet-page"`)
	require.Contains(t, report.GreetEN.Body, `Hello ada`)
	require.Contains(t, report.GreetEN.Body, `Browse the greeting page for ada.`)
	require.Contains(t, report.GreetEN.Body, `rel="canonical" href="https://prefix.example.test/en/greet/ada"`)
	require.Contains(t, report.GreetEN.Body, `hreflang="en" href="https://prefix.example.test/en/greet/ada"`)
	require.Contains(t, report.GreetEN.Body, `hreflang="de" href="https://prefix.example.test/de/greet/ada"`)
	require.Contains(t, report.GreetEN.Body, `id="switch-de" href="https://prefix.example.test/de/greet/ada"`)
	require.NotContains(t, report.GreetEN.Body, `<style type="text/css"`)

	heroClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<section[^>]*data-component="hero"[^>]*>`, report.GreetEN.Body),
	)
	requireClassWithToken(t, report.Stylesheet.Body, heroClasses, "background:#effcf6")

	require.Equal(t, 200, report.GreetDE.Status)
	require.Contains(t, report.GreetDE.Body, `<html lang="de">`)
	require.Contains(t, report.GreetDE.Body, `Hallo ada`)
	require.Contains(t, report.GreetDE.Body, `Begruessungsseite fuer ada lesen.`)
	require.Contains(t, report.GreetDE.Body, `rel="canonical" href="https://prefix.example.test/de/greet/ada"`)
	require.Contains(t, report.GreetDE.Body, `id="switch-en" href="https://prefix.example.test/en/greet/ada"`)

	require.Equal(t, 200, report.GreetPartial.Status)
	require.NotContains(t, report.GreetPartial.Body, `<html`)
	require.Contains(t, report.GreetPartial.Body, `id="greet-page"`)
	require.Contains(t, report.GreetPartial.HXTriggerAfterSettle, `https://prefix.example.test/de/greet/ada`)
	require.Contains(t, report.GreetPartial.HXTriggerAfterSettle, `styles/templ.css`)

	require.Equal(t, 200, report.Stylesheet.Status)
	require.Contains(t, report.Stylesheet.ContentType, "text/css")
	require.Contains(t, report.Stylesheet.Body, "background:#effcf6")
	require.Contains(t, report.Stylesheet.Body, "border:1px solid #17594a")
}

func TestCustomRuntimeFixtureApp(t *testing.T) {
	report := loadCustomRuntimeFixture(t)
	require.Regexp(
		t,
		regexp.MustCompile(`^/build/[0-9a-f]{16}/site\.css$`),
		report.SiteCSSURL,
	)
	require.Regexp(
		t,
		regexp.MustCompile(`^/build/[0-9a-f]{16}/styles/templ\.css$`),
		report.TemplCSSURL,
	)

	require.Equal(t, 200, report.Home.Status)
	require.Contains(t, report.Home.ContentType, "text/html")
	require.Contains(t, report.Home.Body, `id="custom-runtime-page"`)
	require.Contains(t, report.Home.Body, report.SiteCSSURL)
	require.Contains(t, report.Home.Body, report.TemplCSSURL)
	require.Equal(t, "applied", report.Home.XMainMiddleware)
	require.NotContains(t, report.Home.Body, `<style type="text/css"`)

	shellClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<section[^>]*id="runtime-shell"[^>]*>`, report.Home.Body),
	)
	requireClassWithToken(t, report.TemplCSS.Body, shellClasses, "background:#f5f7fb")

	require.Equal(t, 200, report.Extra.Status)
	require.Equal(t, "extra", strings.TrimSpace(report.Extra.Body))
	require.Empty(t, strings.TrimSpace(report.Extra.XMainMiddleware))

	require.Equal(t, 200, report.Public.Status)
	require.Equal(t, "custom-icon", strings.TrimSpace(report.Public.Body))
	require.Empty(t, strings.TrimSpace(report.Public.XMainMiddleware))

	require.Equal(t, 200, report.Health.Status)
	require.Equal(t, "alive", strings.TrimSpace(report.Health.Body))
	require.Empty(t, strings.TrimSpace(report.Health.XMainMiddleware))

	require.Equal(t, 404, report.DefaultHealth.Status)
	require.Contains(t, report.DefaultHealth.Body, `Missing /healthz`)

	require.Equal(t, 200, report.SiteCSS.Status)
	require.Contains(t, report.SiteCSS.ContentType, "text/css")
	require.Contains(t, report.SiteCSS.Body, "font-family:monospace")
	require.Contains(t, report.SiteCSS.Body, "background:#f5f7fb")

	require.Equal(t, 200, report.TemplCSS.Status)
	require.Contains(t, report.TemplCSS.ContentType, "text/css")
	require.Contains(t, report.TemplCSS.Body, "padding:14px")
	require.Contains(t, report.TemplCSS.Body, "border:1px solid #224")
}

func TestTypedModelsFixtureApp(t *testing.T) {
	report := loadTypedModelsFixture(t)

	require.Equal(t, 200, report.Home.Status)
	require.Contains(t, report.Home.Body, `id="typed-root-page"`)
	require.Contains(t, report.Home.Body, `data-heading="Typed root model"`)

	require.Equal(t, 200, report.Marketing.Status)
	require.Contains(t, report.Marketing.Body, `id="marketing-layout"`)
	require.Contains(t, report.Marketing.Body, `data-shell="marketing-shell"`)
	require.Contains(t, report.Marketing.Body, `id="marketing-page"`)
	require.Contains(t, report.Marketing.Body, `data-heading="Typed marketing model"`)
	require.Contains(t, report.Marketing.Body, `id="promo-default"`)
	require.Contains(t, report.Marketing.Body, `data-message="promo-default-model"`)

	require.Equal(t, 404, report.NotFound.Status)
	require.Contains(t, report.NotFound.Body, `id="typed-not-found"`)
	require.Contains(t, report.NotFound.Body, `data-message="typed-not-found-model"`)
	require.Contains(t, report.NotFound.Body, `data-path="/missing"`)

	viewModelsPath := filepath.Join(repoRootPath(t), "e2e", "testdata", "typedmodelsapp", "web", "view", "view_models.go")
	viewModels, err := os.ReadFile(viewModelsPath)
	require.NoError(t, err)
	require.NotContains(t, string(viewModels), "RootLayoutView")
}

func TestTemplRulesFixtureApp(t *testing.T) {
	report := loadTemplRulesFixture(t)
	require.Regexp(
		t,
		regexp.MustCompile(`^/_assets/[0-9a-f]{16}/styles/templ\.css$`),
		report.StylesheetURL,
	)

	t.Run("card", func(t *testing.T) {
		require.Equal(t, 200, report.Card.Status)
		require.Contains(t, report.Card.ContentType, "text/html")
		require.Contains(t, report.Card.Body, `id="card-page"`)
		require.Contains(t, report.Card.Body, `Urgent Card`)
		require.Contains(t, report.Card.Body, report.StylesheetURL)
		require.NotContains(t, report.Card.Body, `<style type="text/css"`)

		cardClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<article[^>]*data-component="card"[^>]*>`, report.Card.Body),
		)
		require.GreaterOrEqual(t, len(cardClasses), 2)
		requireClassWithToken(t, report.TemplCSS.Body, cardClasses, "padding:1rem")
		requireClassWithToken(t, report.TemplCSS.Body, cardClasses, "border-color:#b42318")
	})

	t.Run("panel", func(t *testing.T) {
		require.Equal(t, 200, report.Panel.Status)
		require.Contains(t, report.Panel.Body, `id="panel-page"`)
		require.Contains(t, report.Panel.Body, `id="dock-panel"`)
		require.Contains(t, report.Panel.Body, `data-section="dock"`)
		require.Contains(t, report.Panel.Body, `Component body`)

		panelClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<section[^>]*data-section="dock"[^>]*>`, report.Panel.Body),
		)
		requireClassWithToken(t, report.TemplCSS.Body, panelClasses, "gap:.8rem")
	})

	t.Run("board", func(t *testing.T) {
		require.Equal(t, 200, report.Board.Status)
		require.Contains(t, report.Board.Body, `id="board-page"`)
		require.Contains(t, report.Board.Body, `data-slot="header"`)
		require.Contains(t, report.Board.Body, `Board Header`)
		require.Contains(t, report.Board.Body, `Board Body`)

		boardClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<section[^>]*data-component="board"[^>]*>`, report.Board.Body),
		)
		requireClassWithToken(t, report.TemplCSS.Body, boardClasses, "gap:1rem")
	})

	t.Run("deps", func(t *testing.T) {
		require.Equal(t, 200, report.Deps.Status)
		require.Contains(t, report.Deps.Body, `id="deps-page"`)
		require.Equal(t, 1, strings.Count(report.Deps.Body, `src="/meter.js"`))
		require.Equal(t, 200, report.MeterScript.Status)
		require.Contains(t, report.MeterScript.ContentType, "text/javascript")
		require.Contains(t, report.MeterScript.Body, `export const meter = true;`)
	})

	t.Run("hooks", func(t *testing.T) {
		require.Equal(t, 200, report.Hooks.Status)
		require.Contains(t, report.Hooks.Body, `id="hooks-page"`)
		require.Contains(t, report.Hooks.Body, `x-data="dropdown()"`)
		require.Contains(t, report.Hooks.Body, `x-ref="root"`)
		require.Contains(t, report.Hooks.Body, `data-dropdown-trigger`)
		require.Contains(t, report.Hooks.Body, `x-show="open"`)
		require.Contains(t, report.Hooks.Body, `data-dropdown-panel`)

		hookClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<div[^>]*data-dropdown[^>]*>`, report.Hooks.Body),
		)
		requireClassWithToken(t, report.TemplCSS.Body, hookClasses, "gap:.4rem")
	})

	t.Run("vars", func(t *testing.T) {
		require.Equal(t, 200, report.Vars.Status)
		require.Contains(t, report.Vars.Body, `id="vars-page"`)
		require.Contains(t, report.Vars.Body, `--meter-progress:72%`)
		require.Contains(t, report.Vars.Body, `--meter-accent:#17324d`)

		meterClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<div[^>]*data-meter[^>]*>`, report.Vars.Body),
		)
		requireClassWithToken(t, report.TemplCSS.Body, meterClasses, "inline-size:var(--meter-progress)")
		requireClassWithToken(t, report.TemplCSS.Body, meterClasses, "background:var(--meter-accent)")
	})

	t.Run("fallback", func(t *testing.T) {
		require.Equal(t, 200, report.Fallback.Status)
		require.Contains(t, report.Fallback.Body, `id="fallback-page"`)
		require.Contains(t, report.Fallback.Body, `<style type="text/css">`)
		require.NotContains(t, report.TemplCSS.Body, "width:33%")

		progressClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<div[^>]*data-progress="33"[^>]*>`, report.Fallback.Body),
		)
		require.Len(t, progressClasses, 1)
		requireContainsInlineClassRule(t, report.Fallback.Body, progressClasses[0], "width:33%")
	})

	t.Run("metadata", func(t *testing.T) {
		require.Equal(t, 200, report.Metadata.Status)
		require.Contains(t, report.Metadata.Body, `id="metadata-page"`)
		require.Contains(t, report.Metadata.Body, `<title data-metagen-managed="true">Metadata Variant</title>`)
		require.Contains(t, report.Metadata.Body, `name="description" content="Managed metadata from metagen"`)
		require.Contains(t, report.Metadata.Body, `rel="canonical" href="`+report.BaseURL+`/metadata"`)
		require.Contains(
			t,
			report.Metadata.Body,
			`rel="alternate" type="application/json" href="`+report.BaseURL+`/metadata.json"`,
		)
		require.Contains(t, report.Metadata.Body, `rel="author" href="`+report.BaseURL+`/authors/fixture"`)
		require.Contains(t, report.Metadata.Body, `name="publisher" content="Fixture Publisher"`)
		require.Contains(t, report.Metadata.Body, `name="fixture" content="metadata-variant"`)

		require.Equal(t, 200, report.MetadataPartial.Status)
		require.NotContains(t, report.MetadataPartial.Body, `<html`)
		require.Contains(t, report.MetadataPartial.Body, `id="metadata-page"`)

		var payload map[string]struct {
			Title string `json:"title"`
			Head  string `json:"head"`
		}
		require.NoError(t, json.Unmarshal([]byte(report.MetadataPartial.HXTriggerAfterSettle), &payload))
		patch, ok := payload["metagen:patch"]
		require.True(t, ok)
		require.Equal(t, "Metadata Variant", patch.Title)
		require.NotContains(t, patch.Head, "<title")
		require.Contains(t, patch.Head, `name="description" content="Managed metadata from metagen"`)
		require.Contains(t, patch.Head, `rel="canonical" href="`+report.BaseURL+`/metadata"`)
		require.Contains(t, patch.Head, `name="fixture" content="metadata-variant"`)
		require.Contains(t, patch.Head, `rel="stylesheet" href="`+report.StylesheetURL+`"`)
	})

	t.Run("dashboard", func(t *testing.T) {
		require.Equal(t, 200, report.Dashboard.Status)
		require.Contains(t, report.Dashboard.Body, `data-layout="dashboard"`)
		require.Contains(t, report.Dashboard.Body, `id="dashboard-page"`)
		require.Contains(t, report.Dashboard.Body, `data-slot="inspector"`)
		require.Contains(t, report.Dashboard.Body, `72%`)
		require.NotContains(t, report.Dashboard.Body, `<style type="text/css"`)

		inspectorClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<aside[^>]*data-slot="inspector"[^>]*>`, report.Dashboard.Body),
		)
		requireClassWithToken(t, report.TemplCSS.Body, inspectorClasses, "text-transform:uppercase")

		progressClasses := classesForOpeningTag(
			t,
			mustMatchString(t, `<div[^>]*data-progress="72"[^>]*>`, report.Dashboard.Body),
		)
		require.Len(t, progressClasses, 1)
		requireClassWithToken(t, report.TemplCSS.Body, progressClasses, "width:72%")

		require.Equal(t, 200, report.DashboardPartial.Status)
		require.NotContains(t, report.DashboardPartial.Body, `<html`)
		require.Contains(t, report.DashboardPartial.Body, `id="dashboard-page"`)
		require.NotContains(t, report.DashboardPartial.Body, `data-slot="inspector"`)
	})

	t.Run("stream", func(t *testing.T) {
		require.Equal(t, 200, report.Stream.Status)
		require.Contains(t, report.Stream.ContentType, "text/html")
		require.Equal(t, "first", report.Stream.FirstChunk)
		require.Equal(t, "firstsecond", report.Stream.Body)
	})

	t.Run("stylesheet", func(t *testing.T) {
		require.Equal(t, 200, report.TemplCSS.Status)
		require.Contains(t, report.TemplCSS.ContentType, "text/css")
		require.Contains(t, report.TemplCSS.Body, "padding:1rem")
		require.Contains(t, report.TemplCSS.Body, "inline-size:var(--meter-progress)")
		require.Contains(t, report.TemplCSS.Body, "text-transform:uppercase")
	})
}

func TestExistingTemplgenPathsSkipsMissing(t *testing.T) {
	t.Parallel()

	appDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(appDir, "web", "generated"), 0o755))

	paths := existingTemplgenPaths(t, appDir, "web/generated", "web/components", "web/view")
	require.Equal(t, []string{"web/generated"}, paths)
}
