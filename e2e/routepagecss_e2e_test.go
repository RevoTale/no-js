package e2e

import (
	"net/http"
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

func TestRoutePageCSSConfigFalseDisablesGlobalStylesheet(t *testing.T) {
	bundleConfig := strings.Join([]string{
		"version: 1",
		"assets:",
		"  templ_css: false",
		"",
	}, "\n")
	appDir, server := startPreparedFixtureWithNoJSGenConfig(t, "routepagecssapp", bundleConfig)

	home := requestFixture(t, server, http.MethodGet, "/", nil, requestOptions{})
	partial := requestFixture(t, server, http.MethodGet, "/", nil, hxRequestOptions())

	require.Equal(t, 200, home.Status)
	require.Contains(t, home.Body, `<style type="text/css"`)
	require.NotContains(t, home.Body, `styles/templ.css`)
	require.Contains(t, partial.Body, `<style type="text/css"`)
	require.NotContains(t, partial.HXTriggerAfterSettle, `styles/templ.css`)

	_, err := os.Stat(filepath.Join(appDir, "web/generated", "templ_css_gen.go"))
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Join(appDir, "web/generated", "templcss"))
	require.ErrorIs(t, err, os.ErrNotExist)
}
