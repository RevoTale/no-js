package metagen

import (
	"testing"

	"github.com/RevoTale/no-js/framework/i18n"
)

func BenchmarkBuildAlternates(b *testing.B) {
	cfg := i18n.Config{
		Locales: []string{"en", "uk", "de", "fr", "es"}, DefaultLocale: "en", PrefixMode: i18n.PrefixAsNeeded,
		DisplayLabels: map[string]string{
			"en": "English", "uk": "Українська", "de": "Deutsch", "fr": "Français", "es": "Español",
		},
	}
	b.ReportAllocs()
	for b.Loop() {
		_, err := BuildAlternates("https://example.com/base", cfg, "uk", "/notes?page=2&__live=1", nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
