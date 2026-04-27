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
