package i18n

import (
	"net/http"
	"strings"
)

type RouteDecision struct {
	Locale          string
	OriginalPath    string
	StrippedPath    string
	CanonicalPath   string
	ShouldRedirect  bool
	NotFound        bool
	HadLocalePrefix bool
}

type Resolver struct {
	config Config
}

func NewResolver(cfg Config) (*Resolver, error) {
	normalized, err := NormalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Resolver{config: normalized}, nil
}

func (resolver *Resolver) Config() Config {
	if resolver == nil {
		return Config{}
	}

	return cloneConfig(resolver.config)
}

func (resolver *Resolver) LocaleForRequest(r *http.Request) string {
	requestLocale := ""
	if r != nil {
		requestLocale = LocaleFromContext(r.Context())
	}

	normalized := normalizeLocale(requestLocale)
	if resolver == nil {
		return normalized
	}
	if normalized == "" || !containsLocale(resolver.config, normalized) {
		return resolver.config.DefaultLocale
	}
	return normalized
}

func (resolver *Resolver) LocalizedPath(locale string, pathValue string) string {
	if resolver == nil {
		return NormalizePath(pathValue)
	}
	return localizeNormalizedPath(resolver.config, locale, pathValue)
}

func (resolver *Resolver) Resolve(requestPath string) RouteDecision {
	normalizedPath := NormalizePath(requestPath)
	decision := RouteDecision{
		OriginalPath: normalizedPath,
		StrippedPath: normalizedPath,
	}
	if resolver == nil {
		return decision
	}

	locale, strippedPath, hadPrefix, ok := StripLocale(resolver.config, normalizedPath)
	if !ok {
		return RouteDecision{
			OriginalPath:  normalizedPath,
			StrippedPath:  normalizedPath,
			CanonicalPath: normalizedPath,
			NotFound:      true,
		}
	}

	decision.Locale = locale
	decision.StrippedPath = strippedPath
	decision.HadLocalePrefix = hadPrefix
	decision.CanonicalPath = localizeNormalizedPath(resolver.config, locale, strippedPath)
	decision.ShouldRedirect = decision.CanonicalPath != normalizedPath

	switch resolver.config.PrefixMode {
	case PrefixNever:
		if hadPrefix {
			decision.ShouldRedirect = true
			decision.CanonicalPath = strippedPath
		}
	case PrefixAlways:
		if !hadPrefix {
			decision.ShouldRedirect = true
			decision.CanonicalPath = prefixedPath(locale, strippedPath)
		}
	case PrefixAsNeeded:
		if hadPrefix && strings.EqualFold(locale, resolver.config.DefaultLocale) {
			decision.ShouldRedirect = true
			decision.CanonicalPath = strippedPath
		}
	}

	if strings.TrimSpace(decision.CanonicalPath) == "" {
		decision.CanonicalPath = normalizedPath
	}

	return decision
}

func cloneConfig(cfg Config) Config {
	cloned := Config{
		Locales:       append([]string(nil), cfg.Locales...),
		DefaultLocale: cfg.DefaultLocale,
		PrefixMode:    cfg.PrefixMode,
		DisplayOrder:  append([]string(nil), cfg.DisplayOrder...),
	}
	if len(cfg.DisplayLabels) > 0 {
		cloned.DisplayLabels = make(map[string]string, len(cfg.DisplayLabels))
		for locale, label := range cfg.DisplayLabels {
			cloned.DisplayLabels[locale] = label
		}
	}
	return cloned
}
