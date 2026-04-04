package i18n

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

type runtimeTestKey string

const (
	runtimeTestGreeting runtimeTestKey = "greeting"
	runtimeTestAuthor   runtimeTestKey = "author.description"
)

func TestRuntimeContextProvidesPathsURLsAndLocaleLinks(t *testing.T) {
	t.Parallel()

	runtime, err := NewRuntime(Config{
		Locales:       []string{"en", "de", "fr"},
		DefaultLocale: "en",
		PrefixMode:    PrefixAsNeeded,
		DisplayLabels: map[string]string{
			"en": "English",
			"de": "Deutsch",
		},
		DisplayOrder: []string{"de", "en", "fr"},
	}, nil, map[runtimeTestKey]string{
		runtimeTestGreeting: "Hello",
	})
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodGet, "https://request.example/de/note/hello?page=2", nil)
	request = request.WithContext(WithRequestInfo(request.Context(), RequestInfo{
		Locale:       "de",
		StrippedPath: "/note/hello",
	}))

	root, err := url.Parse("https://example.com/blog")
	require.NoError(t, err)

	i18n := runtime.Context(request, root)
	require.Equal(t, "de", i18n.Locale())
	require.Equal(t, []string{"en", "de", "fr"}, i18n.Locales())
	require.Equal(t, "Hello", i18n.T(runtimeTestGreeting, nil))
	require.Equal(t, "/de/feed.xml", i18n.Path("/feed.xml"))
	require.Equal(t, "/feed.xml", i18n.PathFor("en", "/feed.xml"))
	require.Equal(t, "https://example.com/blog/de/feed.xml", i18n.URL("/feed.xml").String())
	require.Equal(t, "https://example.com/blog/note/hello?page=2", i18n.SwitchURL("en").String())

	links := i18n.LocaleLinks(map[string]string{
		"en": "https://example.com/blog/note/hello?page=2",
		"de": "https://example.com/blog/de/note/hello?page=2",
		"fr": "https://example.com/blog/fr/note/hello?page=2",
	})
	require.Len(t, links, 3)
	require.Equal(t, "de", links[0].Code)
	require.Equal(t, "Deutsch", links[0].Label)
	require.True(t, links[0].Active)
	require.Equal(t, "en", links[1].Code)
	require.Equal(t, "English", links[1].Label)
	require.Equal(t, "fr", links[2].Code)
}

func TestStaticRuntimeLocalizesCompiledMessages(t *testing.T) {
	t.Parallel()

	compiledAuthor, err := CompileMessage("Browse notes by {{.Author}}.")
	require.NoError(t, err)
	compiledAuthorDE, err := CompileMessage("Notizen von {{.Author}}.")
	require.NoError(t, err)

	bundle, err := NewStaticBundle(Config{
		Locales:       []string{"en", "de"},
		DefaultLocale: "en",
		PrefixMode:    PrefixAsNeeded,
	}, map[string]map[runtimeTestKey]CompiledMessage{
		"en": {
			runtimeTestGreeting: {Parts: []CompiledMessagePart{{Text: "Hello"}}},
			runtimeTestAuthor:   compiledAuthor,
		},
		"de": {
			runtimeTestGreeting: {Parts: []CompiledMessagePart{{Text: "Hallo"}}},
			runtimeTestAuthor:   compiledAuthorDE,
		},
	}, map[runtimeTestKey]string{
		runtimeTestGreeting: "Hello",
		runtimeTestAuthor:   "Browse notes by {{.Author}}.",
	})
	require.NoError(t, err)
	require.Equal(t, "en", bundle.Config().DefaultLocale)

	request := httptest.NewRequest(http.MethodGet, "https://example.com/de/", nil)
	request = request.WithContext(WithRequestInfo(request.Context(), RequestInfo{
		Locale:       "de",
		StrippedPath: "/",
	}))

	i18n := bundle.Context(request, nil)
	require.NotNil(t, i18n)
	require.Equal(t, "Hallo", i18n.T(runtimeTestGreeting, nil))
	require.Equal(t, "Notizen von Ada.", i18n.T(runtimeTestAuthor, map[string]any{"Author": "Ada"}))
	require.Equal(t, "Browse notes by Ada.", bundle.Localize("fr", runtimeTestAuthor, map[string]any{"Author": "Ada"}))
}

func TestNewBundleWrapsCatalogRuntime(t *testing.T) {
	t.Parallel()

	bundle, err := NewBundle(Config{
		Locales:       []string{"en", "de"},
		DefaultLocale: "en",
		PrefixMode:    PrefixAsNeeded,
	}, nil, map[runtimeTestKey]string{
		runtimeTestGreeting: "Hello",
	})
	require.NoError(t, err)
	require.Equal(t, "en", bundle.Config().DefaultLocale)
	require.Equal(t, "Hello", bundle.Localize("en", runtimeTestGreeting, nil))
}
