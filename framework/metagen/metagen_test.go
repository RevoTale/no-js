package metagen

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
	"github.com/stretchr/testify/require"
)

func TestHeadRendersManagedSEOAndDeterministicOrder(t *testing.T) {
	t.Parallel()

	first := Metadata{
		Title:       "Example Title",
		Description: "Example Description",
		Alternates: Alternates{
			Canonical: "https://example.com/note/hello",
			Languages: map[string]string{
				"de": "https://example.com/de/note/hello",
				"en": "https://example.com/note/hello",
			},
			Types: map[string]string{
				"application/rss+xml":  "https://example.com/feed.xml",
				"application/atom+xml": "https://example.com/feed.atom",
			},
		},
		Stylesheets: []string{
			"https://example.com/styles/base.css",
			" https://example.com/styles/base.css ",
			"https://example.com/styles/page.css",
		},
		Robots: &Robots{
			Index:  Bool(false),
			Follow: Bool(true),
		},
		OpenGraph: &OpenGraph{
			Type:          "article",
			SiteName:      "Blog",
			Locale:        "en",
			PublishedTime: "2026-03-03T08:00:00Z",
			Authors: []string{
				"https://example.com/authors/b",
				"https://example.com/authors/a",
			},
			Tags: []string{"seo", "framework", "seo"},
			Images: []OpenGraphImage{
				{URL: "https://example.com/images/b.png"},
				{URL: "https://example.com/images/a.png", Alt: "alt-a"},
			},
		},
		Twitter: &Twitter{
			Card:    "summary_large_image",
			Creator: "@example",
			Images: []string{
				"https://example.com/images/z.png",
				"https://example.com/images/a.png",
			},
		},
		Authors: []Author{
			{Name: "Zed", URL: "https://example.com/authors/zed"},
			{Name: "Alice", URL: "https://example.com/authors/alice"},
		},
		Publisher: "Example Publisher",
		Pinterest: &Pinterest{RichPin: Bool(true)},
	}

	second := Metadata{
		Title:       first.Title,
		Description: first.Description,
		Alternates: Alternates{
			Canonical: first.Alternates.Canonical,
			Languages: map[string]string{
				"en": "https://example.com/note/hello",
				"de": "https://example.com/de/note/hello",
			},
			Types: map[string]string{
				"application/atom+xml": "https://example.com/feed.atom",
				"application/rss+xml":  "https://example.com/feed.xml",
			},
		},
		Stylesheets: []string{
			"https://example.com/styles/base.css",
			"https://example.com/styles/page.css",
		},
		Robots: first.Robots,
		OpenGraph: &OpenGraph{
			Type:          "article",
			SiteName:      "Blog",
			Locale:        "en",
			PublishedTime: "2026-03-03T08:00:00Z",
			Authors: []string{
				"https://example.com/authors/a",
				"https://example.com/authors/b",
			},
			Tags: []string{"seo", "framework"},
			Images: []OpenGraphImage{
				{URL: "https://example.com/images/a.png", Alt: "alt-a"},
				{URL: "https://example.com/images/b.png"},
			},
		},
		Twitter: &Twitter{
			Card:    "summary_large_image",
			Creator: "@example",
			Images: []string{
				"https://example.com/images/a.png",
				"https://example.com/images/z.png",
			},
		},
		Authors: []Author{
			{Name: "Alice", URL: "https://example.com/authors/alice"},
			{Name: "Zed", URL: "https://example.com/authors/zed"},
		},
		Publisher: "Example Publisher",
		Pinterest: &Pinterest{RichPin: Bool(true)},
	}

	firstHead := renderHeadToString(t, first)
	secondHead := renderHeadToString(t, second)

	require.Equal(t, firstHead, secondHead)

	required := []string{
		`<title data-metagen-managed="true">Example Title</title>`,
		`name="description" content="Example Description"`,
		`rel="canonical" href="https://example.com/note/hello"`,
		`hreflang="de" href="https://example.com/de/note/hello"`,
		`rel="stylesheet" href="https://example.com/styles/base.css"`,
		`rel="stylesheet" href="https://example.com/styles/page.css"`,
		`property="og:type" content="article"`,
		`property="article:published_time" content="2026-03-03T08:00:00Z"`,
		`property="article:author" content="https://example.com/authors/a"`,
		`property="article:tag" content="framework"`,
		`name="twitter:card" content="summary_large_image"`,
		`name="robots" content="noindex, follow"`,
		`name="author" content="Alice"`,
		`name="pinterest-rich-pin" content="true"`,
	}
	for _, token := range required {
		require.Contains(t, firstHead, token)
	}
}

func TestHeadRendersDangerRawHeadVerbatim(t *testing.T) {
	t.Parallel()

	head := renderHeadToString(t, Metadata{
		Title:         "Raw Head",
		DangerRawHead: []string{`<style id="test-style">.x{color:red}</style>`},
	})

	require.Contains(t, head, `<style id="test-style">.x{color:red}</style>`)
}

func TestMergeAllAppendsDangerRawHeadAndOverridesFields(t *testing.T) {
	t.Parallel()

	parent := Metadata{
		Title:         "Parent",
		Description:   "Parent Description",
		Stylesheets:   []string{" /styles/base.css "},
		DangerRawHead: []string{"<style>.a{}</style>"},
	}
	child := Metadata{
		Title:         "Child",
		Stylesheets:   []string{"/styles/page.css", "/styles/base.css"},
		DangerRawHead: []string{"<script>window.x=1</script>"},
	}

	merged := MergeAll(parent, child)
	require.Equal(t, "Child", merged.Title)
	require.Equal(t, "Parent Description", merged.Description)
	require.Equal(t, []string{"/styles/base.css", "/styles/page.css"}, merged.Stylesheets)
	require.Len(t, merged.DangerRawHead, 2)
	require.Equal(t, "<style>.a{}</style>", merged.DangerRawHead[0])
	require.Equal(t, "<script>window.x=1</script>", merged.DangerRawHead[1])
}

func TestBuildAlternatesPrefixAsNeeded(t *testing.T) {
	t.Parallel()

	alternates, err := BuildAlternates(
		"https://example.com/app",
		frameworki18n.Config{
			Locales:       []string{"en", "de"},
			DefaultLocale: "en",
			PrefixMode:    frameworki18n.PrefixAsNeeded,
		},
		"de",
		"/note/hello?tag=go&__live=navigation",
		map[string]string{
			"application/rss+xml":  "/feed.xml?__live=navigation",
			"application/atom+xml": "https://cdn.example.com/feed.atom?__live=navigation",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "https://example.com/app/de/note/hello?tag=go", alternates.Canonical)
	require.Equal(t, "https://example.com/app/note/hello?tag=go", alternates.Languages["en"])
	require.Equal(t, "https://example.com/app/de/note/hello?tag=go", alternates.Languages["de"])
	require.Equal(t, "https://example.com/app/feed.xml", alternates.Types["application/rss+xml"])
	require.Equal(t, "https://cdn.example.com/feed.atom", alternates.Types["application/atom+xml"])
}

func TestBuildHTMXPatchAndWriteHeaders(t *testing.T) {
	t.Parallel()

	patch, err := BuildHTMXPatch(Metadata{
		Title:       "Notes",
		Description: "A notes feed",
		Stylesheets: []string{"/_assets/hash/styles/templ.css"},
	})
	require.NoError(t, err)
	require.Equal(t, "Notes", patch.Title)
	require.NotContains(t, patch.Head, "<title")
	require.Contains(t, patch.Head, `name="description"`)
	require.Contains(t, patch.Head, `rel="stylesheet" href="/_assets/hash/styles/templ.css"`)

	recorder := httptest.NewRecorder()
	require.NoError(t, WriteHTMXHeaders(recorder, patch))

	rawHeader := recorder.Header().Get("HX-Trigger-After-Settle")
	require.NotEmpty(t, strings.TrimSpace(rawHeader))

	payload := make(map[string]Patch)
	require.NoError(t, json.Unmarshal([]byte(rawHeader), &payload))
	eventPayload, ok := payload[HTMXPatchEvent]
	require.True(t, ok)
	require.Equal(t, "Notes", eventPayload.Title)
}

func TestWriteHTMXHeadersMergesJSONPayload(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	recorder.Header().Set("HX-Trigger-After-Settle", `{"existing":{"ok":true}}`)

	err := WriteHTMXHeaders(recorder, Patch{Title: "Merged"})
	require.NoError(t, err)

	out := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal([]byte(recorder.Header().Get("HX-Trigger-After-Settle")), &out))
	_, ok := out["existing"]
	require.True(t, ok)
	_, ok = out[HTMXPatchEvent]
	require.True(t, ok)
}

func renderHeadToString(t *testing.T, meta Metadata) string {
	t.Helper()

	component := Head(meta)
	var buffer bytes.Buffer
	require.NoError(t, component.Render(context.Background(), &buffer))
	return strings.TrimSpace(buffer.String())
}

func TestWriteHTMXHeadersNilResponseWriter(t *testing.T) {
	t.Parallel()

	var writer http.ResponseWriter
	require.NoError(t, WriteHTMXHeaders(writer, Patch{Title: "noop"}))
}

func TestMergeManagedStylesheetsFromContext(t *testing.T) {
	t.Parallel()

	ctx := WithManagedStylesheets(context.Background(), []string{"/styles/base.css", "/styles/base.css"})
	merged := MergeManagedStylesheets(ctx, Metadata{
		Stylesheets: []string{"/styles/page.css"},
	})

	require.Equal(t, []string{"/styles/base.css", "/styles/page.css"}, merged.Stylesheets)
}

func TestHeadRendersModuleScriptsAfterStylesheets(t *testing.T) {
	t.Parallel()

	head := renderHeadToString(t, Metadata{
		Stylesheets:   []string{"/assets/page.css"},
		ModuleScripts: []string{" /assets/page.js ", "/assets/page.js"},
	})

	stylesheet := `<link data-metagen-managed="true" rel="stylesheet" href="/assets/page.css">`
	script := `<script data-metagen-managed="true" type="module" src="/assets/page.js"></script>`
	require.Contains(t, head, stylesheet)
	require.Contains(t, head, script)
	require.Less(t, strings.Index(head, stylesheet), strings.Index(head, script))
	require.Equal(t, 1, strings.Count(head, script))
}

func TestMergeManagedClientAssetsResolvesStaticURLs(t *testing.T) {
	t.Parallel()

	ctx := WithAssetBasePath(context.Background(), "/_assets/hash")
	merged := MergeManagedClientAssets(ctx, Metadata{
		Stylesheets:   []string{"/global.css"},
		ModuleScripts: []string{"https://cdn.example.com/app.js"},
	}, ClientAssets{
		Stylesheets:   []string{"routes/dashboard.css", "routes/dashboard.css"},
		ModuleScripts: []string{"routes/dashboard.js"},
	})

	require.Equal(t, []string{"/global.css", "/_assets/hash/routes/dashboard.css"}, merged.Stylesheets)
	require.Equal(t, []string{
		"https://cdn.example.com/app.js",
		"/_assets/hash/routes/dashboard.js",
	}, merged.ModuleScripts)
}
