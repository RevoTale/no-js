package i18n

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalizePathAndStripLocale(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Locales:       []string{"en", "uk", "de"},
		DefaultLocale: "en",
		PrefixMode:    PrefixAsNeeded,
	}

	require.Equal(t, "/note/hello", LocalizePath(cfg, "en", "/note/hello"))
	require.Equal(t, "/uk/note/hello", LocalizePath(cfg, "uk", "/note/hello"))

	locale, stripped, hadPrefix, ok := StripLocale(cfg, "/uk/note/hello")
	require.True(t, ok)
	require.Equal(t, "uk", locale)
	require.Equal(t, "/note/hello", stripped)
	require.True(t, hadPrefix)

	_, _, _, ok = StripLocale(cfg, "/it/note/hello")
	require.False(t, ok)
}
