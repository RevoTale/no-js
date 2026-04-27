package e2e

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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
