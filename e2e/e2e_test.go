package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/RevoTale/no-js/internal/filesystem"
	"github.com/stretchr/testify/require"
)

const fixtureModulePath = "example.com/templcssapp"

type probeResponse struct {
	Status               int    `json:"status"`
	Body                 string `json:"body"`
	ContentType          string `json:"content_type"`
	HXTriggerAfterSettle string `json:"hx_trigger_after_settle"`
	Location             string `json:"location"`
}

type probeReport struct {
	Home          probeResponse `json:"home"`
	Partial       probeResponse `json:"partial"`
	NotFound      probeResponse `json:"not_found"`
	Stylesheet    probeResponse `json:"stylesheet"`
	StylesheetURL string        `json:"stylesheet_url"`
}

type namespacedProbeReport struct {
	Dashboard     probeResponse `json:"dashboard"`
	Partial       probeResponse `json:"partial"`
	NotFound      probeResponse `json:"not_found"`
	Stylesheet    probeResponse `json:"stylesheet"`
	StylesheetURL string        `json:"stylesheet_url"`
}

type docsFeatureProbeReport struct {
	AuthorDE                  probeResponse `json:"author_de"`
	AuthorEN                  probeResponse `json:"author_en"`
	AuthorENRedirect          probeResponse `json:"author_en_redirect"`
	AuthorMissing             probeResponse `json:"author_missing"`
	AuthorError               probeResponse `json:"author_error"`
	Dashboard                 probeResponse `json:"dashboard"`
	DashboardPartial          probeResponse `json:"dashboard_partial"`
	PingDE                    probeResponse `json:"ping_de"`
	Robots                    probeResponse `json:"robots"`
	Feed                      probeResponse `json:"feed"`
	AuthorFeed                probeResponse `json:"author_feed"`
	Sitemap                   probeResponse `json:"sitemap"`
	SitemapIndex              probeResponse `json:"sitemap_index"`
	SitemapChunk              probeResponse `json:"sitemap_chunk"`
	Favicon                   probeResponse `json:"favicon"`
	SiteCSS                   probeResponse `json:"site_css"`
	TemplCSS                  probeResponse `json:"templ_css"`
	Health                    probeResponse `json:"health"`
	SiteCSSURL                string        `json:"site_css_url"`
	TemplCSSURL               string        `json:"templ_css_url"`
	ExpensiveLoadsAfterFirst  int           `json:"expensive_loads_after_first"`
	ExpensiveLoadsAfterSecond int           `json:"expensive_loads_after_second"`
}

func TestTemplCSSFixtureApp(t *testing.T) {
	appDir := prepareFixtureApp(t, "templcssapp")
	output := runGo(t, appDir, "run", "./cmd/probe")

	var report probeReport
	require.NoError(t, json.Unmarshal(output, &report))

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
	appDir := prepareFixtureApp(t, "namespacedtemplcssapp")
	output := runGo(t, appDir, "run", "./cmd/probe")

	var report namespacedProbeReport
	require.NoError(t, json.Unmarshal(output, &report))

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
	appDir := prepareFixtureApp(t, "docsfeatureapp")
	output := runGo(t, appDir, "run", "./cmd/probe")

	var report docsFeatureProbeReport
	require.NoError(t, json.Unmarshal(output, &report))

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

func prepareFixtureApp(t *testing.T, fixtureName string) string {
	t.Helper()

	repoRoot := repoRootPath(t)
	appDir := filepath.Join(t.TempDir(), fixtureName)

	require.NoError(
		t,
		filesystem.CopyTree(filepath.Join(repoRoot, "e2e", "testdata", fixtureName), appDir),
	)
	writeFixtureModuleFiles(t, repoRoot, appDir)

	runGo(t, appDir, "run", "github.com/RevoTale/no-js/cmd/no-js", "gen", "routes", "-root", ".")
	runGo(
		t,
		appDir,
		"run",
		"github.com/RevoTale/no-js/cmd/templgen",
		"-base",
		".",
		"-path",
		"web/routes",
		"-path",
		"web/generated",
		"-path",
		"web/components",
	)
	runGo(t, appDir, "run", "github.com/RevoTale/no-js/cmd/no-js", "gen", "assets", "-root", ".", "-templ-css")

	return appDir
}

func writeFixtureModuleFiles(t *testing.T, repoRoot string, appDir string) {
	t.Helper()

	rootGoMod, err := os.ReadFile(filepath.Join(repoRoot, "go.mod"))
	require.NoError(t, err)
	goSum, err := os.ReadFile(filepath.Join(repoRoot, "go.sum"))
	require.NoError(t, err)

	goMod := string(rootGoMod)
	goMod = strings.Replace(goMod, "module github.com/RevoTale/no-js", "module "+fixtureModulePath, 1)
	goMod = strings.TrimSpace(goMod) + "\n\nrequire github.com/RevoTale/no-js v0.0.0\n" +
		"replace github.com/RevoTale/no-js => " + filepath.ToSlash(repoRoot) + "\n"

	require.NoError(t, os.WriteFile(filepath.Join(appDir, "go.mod"), []byte(goMod), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "go.sum"), goSum, 0o644))
}

func runGo(t *testing.T, dir string, args ...string) []byte {
	t.Helper()

	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	goCacheDir := filepath.Join(dir, ".cache", "go-build")
	require.NoError(t, os.MkdirAll(goCacheDir, 0o755))
	cmd.Env = append(
		os.Environ(),
		"GOWORK=off",
		"GOCACHE="+goCacheDir,
	)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", strings.TrimSpace(string(output)))
	return output
}

func repoRootPath(t *testing.T) string {
	t.Helper()

	_, fileName, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(fileName), ".."))
}
