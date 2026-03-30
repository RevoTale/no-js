package i18n

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeConfig(t *testing.T) {
	t.Parallel()

	cfg, err := NormalizeConfig(Config{
		Locales:       []string{"EN", "uk", "uk", "de"},
		DefaultLocale: "EN",
		PrefixMode:    PrefixAsNeeded,
	})
	require.NoError(t, err)
	require.Equal(t, "en", cfg.DefaultLocale)
	require.Len(t, cfg.Locales, 3)
	require.Equal(t, PrefixAsNeeded, cfg.PrefixMode)
}

func TestNormalizeConfigInvalid(t *testing.T) {
	t.Parallel()

	if _, err := NormalizeConfig(Config{
		Locales:       []string{"en"},
		DefaultLocale: "en",
		PrefixMode:    PrefixMode("invalid"),
	}); err == nil {
		require.FailNow(t, "expected invalid prefix mode error")
	}

	if _, err := NormalizeConfig(Config{
		Locales:       []string{"en", "broken-locale"},
		DefaultLocale: "en",
		PrefixMode:    PrefixAsNeeded,
	}); err == nil {
		require.FailNow(t, "expected invalid locale error")
	}
}
