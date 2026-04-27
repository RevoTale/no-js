package e2e

import (
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientAssetsFixtureAppIncludesGroupedSlotAndNestedLayoutAssets(t *testing.T) {
	appDir, server := startPreparedFixture(t, "clientassetsslotgroupapp")
	ops := requestFixture(t, server, http.MethodGet, "/ops", nil, requestOptions{})
	reports := requestFixture(t, server, http.MethodGet, "/ops/reports", nil, requestOptions{})

	rootCSSURL := extractPatternURL(t, routeRootCSSPattern, ops.Body, "root stylesheet")
	opsLayoutCSSURL := extractPatternURL(
		t,
		regexp.MustCompile(`href="([^"]+/routes/_group__lab/_group__experiments/ops/layout\.css)"`),
		ops.Body,
		"grouped ops layout stylesheet",
	)
	opsLayoutScriptURL := extractPatternURL(
		t,
		regexp.MustCompile(`src="([^"]+/routes/_group__lab/_group__experiments/ops/layout\.js)"`),
		ops.Body,
		"grouped ops layout script",
	)
	opsPageScriptURL := extractPatternURL(
		t,
		regexp.MustCompile(`src="([^"]+/routes/_group__lab/_group__experiments/ops/page\.js)"`),
		ops.Body,
		"grouped ops page script",
	)
	opsSlotLayoutScriptURL := extractPatternURL(
		t,
		regexp.MustCompile(`src="([^"]+/routes/_group__lab/_group__experiments/ops/_slot__panel/layout\.js)"`),
		ops.Body,
		"grouped ops slot layout script",
	)
	opsSlotScriptURL := extractPatternURL(
		t,
		regexp.MustCompile(`src="([^"]+/routes/_group__lab/_group__experiments/ops/_slot__panel/default\.js)"`),
		ops.Body,
		"grouped ops slot default script",
	)
	insightScriptURL := extractPatternURL(
		t,
		regexp.MustCompile(`src="([^"]+/components/insight/insight\.js)"`),
		ops.Body,
		"insight component script",
	)

	reportsOpsLayoutCSSURL := extractPatternURL(
		t,
		regexp.MustCompile(`href="([^"]+/routes/_group__lab/_group__experiments/ops/layout\.css)"`),
		reports.Body,
		"grouped ops layout stylesheet on reports page",
	)
	reportsLayoutCSSURL := extractPatternURL(
		t,
		regexp.MustCompile(`href="([^"]+/routes/_group__lab/_group__experiments/ops/reports/layout\.css)"`),
		reports.Body,
		"grouped reports layout stylesheet",
	)
	reportsLayoutScriptURL := extractPatternURL(
		t,
		regexp.MustCompile(`src="([^"]+/routes/_group__lab/_group__experiments/ops/reports/layout\.js)"`),
		reports.Body,
		"grouped reports layout script",
	)
	reportsPageScriptURL := extractPatternURL(
		t,
		regexp.MustCompile(`src="([^"]+/routes/_group__lab/_group__experiments/ops/reports/page\.js)"`),
		reports.Body,
		"grouped reports page script",
	)
	reportsSlotLayoutScriptURL := extractPatternURL(
		t,
		regexp.MustCompile(`src="([^"]+/routes/_group__lab/_group__experiments/ops/_slot__panel/layout\.js)"`),
		reports.Body,
		"grouped ops slot layout script on reports page",
	)
	reportsSlotReportsLayoutScriptURL := extractPatternURL(
		t,
		regexp.MustCompile(
			`src="([^"]+/routes/_group__lab/_group__experiments/ops/_slot__panel/_group__parallel/reports/layout\.js)"`,
		),
		reports.Body,
		"grouped ops slot reports layout script",
	)
	reportsSlotReportsPageScriptURL := extractPatternURL(
		t,
		regexp.MustCompile(
			`src="([^"]+/routes/_group__lab/_group__experiments/ops/_slot__panel/_group__parallel/reports/page\.js)"`,
		),
		reports.Body,
		"grouped ops slot reports page script",
	)
	reportsInsightScriptURL := extractPatternURL(
		t,
		regexp.MustCompile(`src="([^"]+/components/insight/insight\.js)"`),
		reports.Body,
		"insight component script on reports page",
	)

	require.Equal(t, reportsOpsLayoutCSSURL, opsLayoutCSSURL)
	require.Equal(t, reportsSlotLayoutScriptURL, opsSlotLayoutScriptURL)
	require.Equal(t, reportsInsightScriptURL, insightScriptURL)

	require.Equal(t, 200, ops.Status)
	require.Contains(t, ops.Body, `data-layout="grouped-ops"`)
	require.Contains(t, ops.Body, `data-slot-layout="ops-panel"`)
	require.Contains(t, ops.Body, `data-slot="ops-panel-default"`)
	require.Contains(t, ops.Body, `data-route-page="grouped-ops"`)
	require.Contains(t, ops.Body, `data-component="insight"`)
	requireAssetOrder(t, ops.Body, rootCSSURL, opsLayoutCSSURL)
	require.Equal(t, 1, strings.Count(ops.Body, opsLayoutCSSURL))
	require.Equal(t, 1, strings.Count(ops.Body, opsLayoutScriptURL))
	require.Equal(t, 1, strings.Count(ops.Body, opsSlotLayoutScriptURL))
	require.Equal(t, 1, strings.Count(ops.Body, opsSlotScriptURL))
	require.Equal(t, 1, strings.Count(ops.Body, opsPageScriptURL))
	require.Equal(t, 1, strings.Count(ops.Body, insightScriptURL))
	require.NotContains(t, ops.Body, `/routes/_group__lab/_group__experiments/ops/page.css`)
	require.NotContains(t, ops.Body, `/routes/_group__lab/_group__experiments/ops/_slot__panel/layout.css`)
	require.NotContains(t, ops.Body, `/routes/_group__lab/_group__experiments/ops/_slot__panel/default.css`)
	require.NotContains(t, ops.Body, `/components/insight/insight.css`)
	require.NotContains(t, ops.Body, `/routes/_group__lab/_group__experiments/ops/reports/layout.css`)
	require.NotContains(t, ops.Body, `/routes/_group__lab/_group__experiments/ops/reports/page.js`)
	require.Equal(t, 1, strings.Count(ops.Body, reportsSlotReportsLayoutScriptURL))
	require.Equal(t, 1, strings.Count(ops.Body, reportsSlotReportsPageScriptURL))

	require.Equal(t, 200, reports.Status)
	require.Contains(t, reports.Body, `data-layout="grouped-ops"`)
	require.Contains(t, reports.Body, `data-layout="grouped-ops-reports"`)
	require.Contains(t, reports.Body, `data-slot-layout="ops-panel"`)
	require.Contains(t, reports.Body, `data-slot-layout="ops-panel-reports"`)
	require.Contains(t, reports.Body, `data-slot="ops-panel-reports"`)
	require.Contains(t, reports.Body, `data-route-page="grouped-ops-reports"`)
	require.Contains(t, reports.Body, `data-component="insight"`)
	requireAssetOrder(
		t,
		reports.Body,
		rootCSSURL,
		opsLayoutCSSURL,
		reportsLayoutCSSURL,
	)
	require.Equal(t, 1, strings.Count(reports.Body, opsLayoutCSSURL))
	require.Equal(t, 1, strings.Count(reports.Body, reportsLayoutCSSURL))
	require.Equal(t, 1, strings.Count(reports.Body, opsLayoutScriptURL))
	require.Equal(t, 1, strings.Count(reports.Body, opsSlotLayoutScriptURL))
	require.Equal(t, 1, strings.Count(reports.Body, reportsSlotReportsLayoutScriptURL))
	require.Equal(t, 1, strings.Count(reports.Body, reportsSlotReportsPageScriptURL))
	require.Equal(t, 1, strings.Count(reports.Body, reportsLayoutScriptURL))
	require.Equal(t, 1, strings.Count(reports.Body, reportsPageScriptURL))
	require.Equal(t, 1, strings.Count(reports.Body, insightScriptURL))
	require.NotContains(t, reports.Body, opsPageScriptURL)
	require.Equal(t, 1, strings.Count(reports.Body, opsSlotScriptURL))
	require.NotContains(t, reports.Body, `/routes/_group__lab/_group__experiments/ops/reports/page.css`)
	require.NotContains(t, reports.Body, `/routes/_group__lab/_group__experiments/ops/_slot__panel/layout.css`)
	require.NotContains(t, reports.Body, `/routes/_group__lab/_group__experiments/ops/_slot__panel/default.css`)
	require.NotContains(
		t,
		reports.Body,
		`/routes/_group__lab/_group__experiments/ops/_slot__panel/_group__parallel/reports/layout.css`,
	)
	require.NotContains(
		t,
		reports.Body,
		`/routes/_group__lab/_group__experiments/ops/_slot__panel/_group__parallel/reports/page.css`,
	)
	require.NotContains(t, reports.Body, `/components/insight/insight.css`)

	opsLayoutCSS := requestFixture(t, server, http.MethodGet, opsLayoutCSSURL, nil, requestOptions{})
	reportsLayoutCSS := requestFixture(t, server, http.MethodGet, reportsLayoutCSSURL, nil, requestOptions{})
	opsLayoutScript := requestFixture(t, server, http.MethodGet, opsLayoutScriptURL, nil, requestOptions{})
	opsPageScript := requestFixture(t, server, http.MethodGet, opsPageScriptURL, nil, requestOptions{})
	opsSlotLayoutScript := requestFixture(t, server, http.MethodGet, opsSlotLayoutScriptURL, nil, requestOptions{})
	opsSlotScript := requestFixture(t, server, http.MethodGet, opsSlotScriptURL, nil, requestOptions{})
	reportsLayoutScript := requestFixture(t, server, http.MethodGet, reportsLayoutScriptURL, nil, requestOptions{})
	reportsPageScript := requestFixture(t, server, http.MethodGet, reportsPageScriptURL, nil, requestOptions{})
	reportsSlotReportsLayoutScript := requestFixture(
		t,
		server,
		http.MethodGet,
		reportsSlotReportsLayoutScriptURL,
		nil,
		requestOptions{},
	)
	reportsSlotReportsPageScript := requestFixture(
		t,
		server,
		http.MethodGet,
		reportsSlotReportsPageScriptURL,
		nil,
		requestOptions{},
	)
	insightScript := requestFixture(t, server, http.MethodGet, insightScriptURL, nil, requestOptions{})

	require.Equal(t, 200, opsLayoutCSS.Status)
	require.Contains(t, opsLayoutCSS.ContentType, "text/css")
	opsLayoutClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<section[^>]*data-layout="grouped-ops"[^>]*>`, ops.Body),
	)
	opsPageClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<article[^>]*data-route-page="grouped-ops"[^>]*>`, ops.Body),
	)
	opsInsightClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<section[^>]*data-component="insight"[^>]*>`, ops.Body),
	)
	requireClassWithToken(t, opsLayoutCSS.Body, opsLayoutClasses, "--ops-layout-css")
	requireClassWithToken(t, opsLayoutCSS.Body, opsPageClasses, "--ops-page-css")
	requireClassWithToken(t, opsLayoutCSS.Body, opsInsightClasses, "--insight-component-css")
	opsPanelLayoutClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<section[^>]*data-slot-layout="ops-panel"[^>]*>`, ops.Body),
	)
	opsPanelDefaultClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<aside[^>]*data-slot="ops-panel-default"[^>]*>`, ops.Body),
	)
	requireClassWithToken(t, opsLayoutCSS.Body, opsPanelLayoutClasses, "--ops-slot-layout-css")
	requireClassWithToken(t, opsLayoutCSS.Body, opsPanelDefaultClasses, "--ops-slot-default-css")
	require.Contains(t, opsLayoutCSS.Body, "--ops-slot-reports-layout-css")
	require.Contains(t, opsLayoutCSS.Body, "--ops-slot-reports-page-css")
	require.NotContains(t, opsLayoutCSS.Body, "--ops-reports-layout-css")
	require.NotContains(t, opsLayoutCSS.Body, "--ops-reports-page-css")
	requireNoOriginalClassSelector(t, opsLayoutCSS.Body, "shell")
	requireNoOriginalClassSelector(t, opsLayoutCSS.Body, "panel")
	requireNoOriginalClassSelector(t, opsLayoutCSS.Body, "root")

	require.Equal(t, 200, reportsLayoutCSS.Status)
	require.Contains(t, reportsLayoutCSS.ContentType, "text/css")
	reportsLayoutClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<section[^>]*data-layout="grouped-ops-reports"[^>]*>`, reports.Body),
	)
	reportsPageClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<article[^>]*data-route-page="grouped-ops-reports"[^>]*>`, reports.Body),
	)
	reportsInsightClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<section[^>]*data-component="insight"[^>]*>`, reports.Body),
	)
	requireClassWithToken(t, reportsLayoutCSS.Body, reportsLayoutClasses, "--ops-reports-layout-css")
	requireClassWithToken(t, reportsLayoutCSS.Body, reportsPageClasses, "--ops-reports-page-css")
	requireClassWithToken(t, reportsLayoutCSS.Body, reportsInsightClasses, "--insight-component-css")
	require.NotContains(t, reportsLayoutCSS.Body, "--ops-page-css")
	require.NotContains(t, reportsLayoutCSS.Body, "--ops-slot-default-css")
	require.NotContains(t, reportsLayoutCSS.Body, "--ops-slot-reports-page-css")
	reportsPanelReportsLayoutClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<section[^>]*data-slot-layout="ops-panel-reports"[^>]*>`, reports.Body),
	)
	reportsPanelReportsPageClasses := classesForOpeningTag(
		t,
		mustMatchString(t, `<aside[^>]*data-slot="ops-panel-reports"[^>]*>`, reports.Body),
	)
	requireClassWithToken(t, opsLayoutCSS.Body, reportsPanelReportsLayoutClasses, "--ops-slot-reports-layout-css")
	requireClassWithToken(t, opsLayoutCSS.Body, reportsPanelReportsPageClasses, "--ops-slot-reports-page-css")

	require.Equal(t, 200, opsLayoutScript.Status)
	require.Contains(t, opsLayoutScript.Body, "clientAssetsOpsLayout")
	require.NotContains(t, opsLayoutScript.Body, "clientAssetsOpsPage")
	require.Equal(t, 200, opsPageScript.Status)
	require.Contains(t, opsPageScript.Body, "clientAssetsOpsPage")
	require.NotContains(t, opsPageScript.Body, "clientAssetsOpsReportsPage")
	require.Equal(t, 200, opsSlotLayoutScript.Status)
	require.Contains(t, opsSlotLayoutScript.Body, "clientAssetsOpsSlotLayout")
	require.Equal(t, 200, opsSlotScript.Status)
	require.Contains(t, opsSlotScript.Body, "clientAssetsOpsSlotDefault")
	require.Equal(t, 200, reportsLayoutScript.Status)
	require.Contains(t, reportsLayoutScript.Body, "clientAssetsOpsReportsLayout")
	require.Equal(t, 200, reportsPageScript.Status)
	require.Contains(t, reportsPageScript.Body, "clientAssetsOpsReportsPage")
	require.NotContains(t, reportsPageScript.Body, "clientAssetsOpsPage")
	require.Equal(t, 200, reportsSlotReportsLayoutScript.Status)
	require.Contains(t, reportsSlotReportsLayoutScript.Body, "clientAssetsOpsSlotReportsLayout")
	require.Equal(t, 200, reportsSlotReportsPageScript.Status)
	require.Contains(t, reportsSlotReportsPageScript.Body, "clientAssetsOpsSlotReportsPage")
	require.Equal(t, 200, insightScript.Status)
	require.Contains(t, insightScript.Body, "clientAssetsInsight")

	require.FileExists(
		t,
		filepath.Join(appDir, "web", "routes", "_group__lab", "_group__experiments", "ops", "layout.css_gen.go"),
	)
	require.FileExists(
		t,
		filepath.Join(appDir, "web", "routes", "_group__lab", "_group__experiments", "ops", "page.css_gen.go"),
	)
	require.FileExists(
		t,
		filepath.Join(
			appDir,
			"web",
			"routes",
			"_group__lab",
			"_group__experiments",
			"ops",
			"_slot__panel",
			"layout.css_gen.go",
		),
	)
	require.FileExists(
		t,
		filepath.Join(
			appDir,
			"web",
			"routes",
			"_group__lab",
			"_group__experiments",
			"ops",
			"_slot__panel",
			"default.css_gen.go",
		),
	)
	require.FileExists(
		t,
		filepath.Join(
			appDir,
			"web",
			"routes",
			"_group__lab",
			"_group__experiments",
			"ops",
			"reports",
			"layout.css_gen.go",
		),
	)
	require.FileExists(
		t,
		filepath.Join(
			appDir,
			"web",
			"routes",
			"_group__lab",
			"_group__experiments",
			"ops",
			"reports",
			"page.css_gen.go",
		),
	)
	require.FileExists(
		t,
		filepath.Join(
			appDir,
			"web",
			"routes",
			"_group__lab",
			"_group__experiments",
			"ops",
			"_slot__panel",
			"_group__parallel",
			"reports",
			"layout.css_gen.go",
		),
	)
	require.FileExists(
		t,
		filepath.Join(
			appDir,
			"web",
			"routes",
			"_group__lab",
			"_group__experiments",
			"ops",
			"_slot__panel",
			"_group__parallel",
			"reports",
			"page.css_gen.go",
		),
	)
	require.FileExists(t, filepath.Join(appDir, "web", "components", "insight", "insight.css_gen.go"))
	require.FileExists(t, filepath.Join(appDir, "web", "components", "insight", "insight.ts_gen.go"))
}

func requireAssetOrder(t *testing.T, html string, urls ...string) {
	t.Helper()

	previous := -1
	for _, url := range urls {
		index := strings.Index(html, url)
		require.NotEqual(t, -1, index, "asset %q not found", url)
		require.Greater(t, index, previous, "asset %q is out of order", url)
		previous = index
	}
}
