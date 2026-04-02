package i18n

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

type runtimeTestKey string

const runtimeTestGreeting runtimeTestKey = "greeting"

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
