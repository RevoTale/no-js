package webi18n

import frameworki18n "github.com/RevoTale/no-js/framework/i18n"

func Config() frameworki18n.Config {
	return frameworki18n.Config{
		Locales:       []string{"en", "de"},
		DefaultLocale: "en",
		PrefixMode:    frameworki18n.PrefixAsNeeded,
		DisplayLabels: map[string]string{
			"en": "English",
			"de": "Deutsch",
		},
		DisplayOrder: []string{"de", "en"},
	}
}
