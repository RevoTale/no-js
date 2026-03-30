package i18n

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverResolveAsNeeded(t *testing.T) {
	t.Parallel()

	resolver, err := NewResolver(Config{
		Locales:       []string{"en", "uk", "de"},
		DefaultLocale: "en",
		PrefixMode:    PrefixAsNeeded,
	})
	require.NoError(t, err)

	root := resolver.Resolve("/")
	require.Equal(t, "en", root.Locale)
	require.Equal(t, "/", root.StrippedPath)
	require.False(t, root.ShouldRedirect)

	localized := resolver.Resolve("/uk/note/hello")
	require.Equal(t, "uk", localized.Locale)
	require.Equal(t, "/note/hello", localized.StrippedPath)
	require.False(t, localized.ShouldRedirect)

	defaultPrefixed := resolver.Resolve("/en/note/hello")
	require.True(t, defaultPrefixed.ShouldRedirect)
	require.Equal(t, "/note/hello", defaultPrefixed.CanonicalPath)

	unknown := resolver.Resolve("/it/note/hello")
	require.True(t, unknown.NotFound)
}
