package e2e

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientAssetsFixtureApp(t *testing.T) {
	report := loadClientAssetsFixture(t)

	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/root\.css$`), report.RootCSSURL)
	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/site\.css$`), report.SiteCSSURL)
	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/page\.css$`), report.RouteCSSURL)
	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/page\.js$`), report.RouteScriptURL)
	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/components/meter/meter\.js$`), report.MeterScriptURL)
	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/section/layout\.css$`), report.SectionCSSURL)
	require.Regexp(
		t,
		regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/section/layout\.js$`),
		report.SectionLayoutScriptURL,
	)
	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/section/page\.js$`), report.SectionPageScriptURL)
	require.Regexp(
		t,
		regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/section/admin/layout\.css$`),
		report.SectionAdminCSSURL,
	)
	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/complex/page\.css$`), report.ComplexCSSURL)
	require.Regexp(t, regexp.MustCompile(`^/_assets/[0-9a-f]{16}/routes/404\.css$`), report.NotFoundCSSURL)

	require.Equal(t, 200, report.Home.Status)
	require.Contains(t, report.Home.Body, `data-route-page="home"`)
	require.Contains(t, report.Home.Body, report.RootCSSURL)
	require.Contains(t, report.Home.Body, report.SiteCSSURL)
	require.Contains(t, report.Home.Body, report.RouteCSSURL)
	require.Contains(t, report.Home.Body, report.RouteScriptURL)
	require.Contains(t, report.Home.Body, report.MeterScriptURL)
	require.Equal(t, 1, strings.Count(report.Home.Body, report.RootCSSURL))
	require.Equal(t, 1, strings.Count(report.Home.Body, report.SiteCSSURL))
	require.Equal(t, 1, strings.Count(report.Home.Body, report.RouteCSSURL))
	require.Equal(t, 1, strings.Count(report.Home.Body, report.RouteScriptURL))
	require.Equal(t, 1, strings.Count(report.Home.Body, report.MeterScriptURL))
	require.NotContains(t, report.Home.Body, report.SectionCSSURL)
	require.NotContains(t, report.Home.Body, report.SectionLayoutScriptURL)
	require.NotContains(t, report.Home.Body, report.SectionPageScriptURL)
	require.NotContains(t, report.Home.Body, report.ComplexCSSURL)
	require.NotContains(t, report.Home.Body, "styles/templ.css")
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

	require.Equal(t, 200, report.SiteCSS.Status)
	require.Contains(t, report.SiteCSS.ContentType, "text/css")
	require.Contains(t, report.SiteCSS.Body, ".client-assets-imported-reset")
	require.Contains(t, report.SiteCSS.Body, "--client-assets-imported-reset")
	require.Contains(t, report.SiteCSS.Body, ".client-assets-imported-icon")
	require.Contains(t, report.SiteCSS.Body, "url(./icons/mark.svg)")
	require.Contains(t, report.SiteCSS.Body, ".client-assets-site")
	require.Less(
		t,
		strings.Index(report.SiteCSS.Body, ".client-assets-imported-reset"),
		strings.Index(report.SiteCSS.Body, ".client-assets-site"),
	)
	require.NotContains(t, report.SiteCSS.Body, "@import")

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

	require.Equal(t, 200, report.RootCSS.Status)
	require.Contains(t, report.RootCSS.ContentType, "text/css")
	require.Contains(t, report.RootCSS.Body, "--client-assets-root-shell")
	require.Contains(t, report.RootCSS.Body, "-webkit-user-select:none")
	require.Contains(t, report.RootCSS.Body, "user-select:none")
	require.NotContains(t, report.RootCSS.Body, "padding:18px")
	require.NotContains(t, report.RootCSS.Body, "--section-summary-page")

	require.Equal(t, 200, report.RouteScript.Status)
	require.Contains(t, report.RouteScript.ContentType, "text/javascript")
	require.Contains(t, report.RouteScript.Body, "clientAssetsPage")
	require.NotContains(t, report.RouteScript.Body, "clientAssetsMeter")
	require.NotContains(t, report.RouteScript.Body, "clientAssetsSectionLayout")
	require.NotContains(t, report.RouteScript.Body, "clientAssetsSectionPage")

	require.Equal(t, 200, report.MeterScript.Status)
	require.Contains(t, report.MeterScript.ContentType, "text/javascript")
	require.Contains(t, report.MeterScript.Body, "clientAssetsMeter")
	require.NotContains(t, report.MeterScript.Body, "clientAssetsPage")

	require.Equal(t, 200, report.About.Status)
	require.Contains(t, report.About.Body, `data-route-page="about"`)
	require.Contains(t, report.About.Body, report.RootCSSURL)
	require.NotContains(t, report.About.Body, `/routes/page.css`)
	require.NotContains(t, report.About.Body, `/routes/page.js`)
	require.NotContains(t, report.About.Body, `/routes/section/layout.css`)
	require.NotContains(t, report.About.Body, `/routes/section/layout.js`)
	require.NotContains(t, report.About.Body, `/routes/complex/page.css`)

	require.Equal(t, 200, report.Section.Status)
	require.Contains(t, report.Section.Body, `data-layout="section"`)
	require.Contains(t, report.Section.Body, `data-route-page="section"`)
	require.Contains(t, report.Section.Body, report.RootCSSURL)
	require.Contains(t, report.Section.Body, report.SectionCSSURL)
	require.Contains(t, report.Section.Body, report.SectionLayoutScriptURL)
	require.Contains(t, report.Section.Body, report.SectionPageScriptURL)
	require.Contains(t, report.Section.Body, report.MeterScriptURL)
	require.Equal(t, 1, strings.Count(report.Section.Body, report.SectionCSSURL))
	require.Equal(t, 1, strings.Count(report.Section.Body, report.SectionLayoutScriptURL))
	require.Equal(t, 1, strings.Count(report.Section.Body, report.SectionPageScriptURL))
	require.NotContains(t, report.Section.Body, report.RouteCSSURL)
	require.NotContains(t, report.Section.Body, `/components/meter/meter.css`)
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
	require.Contains(t, report.SectionCSS.Body, "--section-summary-page")
	require.NotContains(t, report.SectionCSS.Body, "--section-admin-page")

	require.Equal(t, 200, report.SectionSummary.Status)
	require.Contains(t, report.SectionSummary.Body, `data-route-page="section-summary"`)
	require.Contains(t, report.SectionSummary.Body, report.RootCSSURL)
	require.Contains(t, report.SectionSummary.Body, report.SectionCSSURL)
	require.Contains(t, report.SectionSummary.Body, report.SectionLayoutScriptURL)
	require.Contains(t, report.SectionSummary.Body, report.MeterScriptURL)
	require.NotContains(t, report.SectionSummary.Body, report.SectionAdminCSSURL)
	require.NotContains(t, report.SectionSummary.Body, report.SectionPageScriptURL)

	require.Equal(t, 200, report.SectionAdmin.Status)
	require.Contains(t, report.SectionAdmin.Body, `data-layout="section"`)
	require.Contains(t, report.SectionAdmin.Body, `data-layout="section-admin"`)
	require.Contains(t, report.SectionAdmin.Body, `data-route-page="section-admin"`)
	require.Contains(t, report.SectionAdmin.Body, report.RootCSSURL)
	require.Contains(t, report.SectionAdmin.Body, report.SectionCSSURL)
	require.Contains(t, report.SectionAdmin.Body, report.SectionAdminCSSURL)
	require.NotContains(t, report.SectionAdmin.Body, report.RouteCSSURL)
	require.Equal(t, 200, report.SectionAdminCSS.Status)
	require.Contains(t, report.SectionAdminCSS.ContentType, "text/css")
	require.Contains(t, report.SectionAdminCSS.Body, "--section-admin-layout")
	require.Contains(t, report.SectionAdminCSS.Body, "--section-admin-page")
	require.NotContains(t, report.SectionAdminCSS.Body, "removed by the final esbuild")
	require.NotContains(t, report.SectionAdminCSS.Body, "--section-summary-page")

	require.Equal(t, 200, report.SectionLayoutScript.Status)
	require.Contains(t, report.SectionLayoutScript.ContentType, "text/javascript")
	require.Contains(t, report.SectionLayoutScript.Body, "clientAssetsSectionLayout")
	require.NotContains(t, report.SectionLayoutScript.Body, "clientAssetsSectionPage")
	require.NotContains(t, report.SectionLayoutScript.Body, "clientAssetsPage")

	require.Equal(t, 200, report.SectionPageScript.Status)
	require.Contains(t, report.SectionPageScript.ContentType, "text/javascript")
	require.Contains(t, report.SectionPageScript.Body, "clientAssetsSectionPage")
	require.NotContains(t, report.SectionPageScript.Body, "clientAssetsSectionLayout")
	require.NotContains(t, report.SectionPageScript.Body, "clientAssetsPage")

	requireComplexClientAssetCSS(t, report)

	require.Equal(t, 404, report.NotFound.Status)
	require.Contains(t, report.NotFound.Body, report.RootCSSURL)
	require.Contains(t, report.NotFound.Body, report.NotFoundCSSURL)
	require.Equal(t, 1, strings.Count(report.NotFound.Body, report.NotFoundCSSURL))
	require.NotContains(t, report.NotFound.Body, report.RouteScriptURL)
	require.NotContains(t, report.NotFound.Body, report.SectionLayoutScriptURL)
	require.Equal(t, 200, report.NotFoundCSS.Status)
	require.Contains(t, report.NotFoundCSS.Body, "border:3px dashed #b42318")

	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "page.css_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "page.ts_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "section", "layout.css_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "section", "layout.ts_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "section", "page.css_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "section", "page.ts_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "section", "summary", "page.css_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "section", "admin", "layout.css_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "section", "admin", "page.css_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "routes", "complex", "page.css_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "components", "meter", "meter.css_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "components", "meter", "meter.tsx_gen.go"))
	require.FileExists(t, filepath.Join(report.AppDir, "web", "components", "filterpanel", "filterpanel.css_gen.go"))
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
