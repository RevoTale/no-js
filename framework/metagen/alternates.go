package metagen

import (
	"fmt"
	"net/url"
	"path"
	"slices"
	"strings"

	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
)

func BuildAlternates(
	rootURL string,
	cfg frameworki18n.Config,
	locale string,
	pathWithQuery string,
	types map[string]string,
) (Alternates, error) {
	absoluteRoot, err := parseRootURL(rootURL)
	if err != nil {
		return Alternates{}, err
	}

	normalizedCfg, err := frameworki18n.NormalizeConfig(cfg)
	if err != nil {
		return Alternates{}, fmt.Errorf("normalize i18n config: %w", err)
	}

	normalizedLocale := strings.ToLower(strings.TrimSpace(locale))
	if normalizedLocale == "" || !slices.Contains(normalizedCfg.Locales, normalizedLocale) {
		normalizedLocale = normalizedCfg.DefaultLocale
	}

	pagePath, query, err := parsePathWithQuery(pathWithQuery)
	if err != nil {
		return Alternates{}, err
	}

	canonicalPath := frameworki18n.LocalizePath(normalizedCfg, normalizedLocale, pagePath)
	alternates := Alternates{
		Canonical: absoluteURL(absoluteRoot, canonicalPath, query),
		Languages: make(map[string]string, len(normalizedCfg.Locales)),
	}

	for _, localeCode := range normalizedCfg.Locales {
		localizedPath := frameworki18n.LocalizePath(normalizedCfg, localeCode, pagePath)
		alternates.Languages[localeCode] = absoluteURL(absoluteRoot, localizedPath, query)
	}

	if len(types) > 0 {
		alternates.Types = make(map[string]string, len(types))
		for mediaType, href := range types {
			trimmedType := strings.TrimSpace(mediaType)
			trimmedHref := strings.TrimSpace(href)
			if trimmedType == "" || trimmedHref == "" {
				continue
			}

			absoluteTypeURL, buildErr := buildAbsoluteTypeURL(absoluteRoot, trimmedHref)
			if buildErr != nil {
				return Alternates{}, fmt.Errorf("build alternate type %q: %w", trimmedType, buildErr)
			}
			alternates.Types[trimmedType] = absoluteTypeURL
		}
	}

	return Normalize(Metadata{Alternates: alternates}).Alternates, nil
}

func parseRootURL(rootURL string) (*url.URL, error) {
	trimmedRoot := strings.TrimSpace(rootURL)
	if trimmedRoot == "" {
		return nil, fmt.Errorf("root URL is required")
	}

	parsed, err := url.Parse(trimmedRoot)
	if err != nil {
		return nil, fmt.Errorf("parse root URL %q: %w", trimmedRoot, err)
	}
	if !parsed.IsAbs() || strings.TrimSpace(parsed.Host) == "" {
		return nil, fmt.Errorf("root URL %q must be absolute", trimmedRoot)
	}

	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed, nil
}

func parsePathWithQuery(pathWithQuery string) (string, url.Values, error) {
	trimmed := strings.TrimSpace(pathWithQuery)
	if trimmed == "" {
		return "/", url.Values{}, nil
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", nil, fmt.Errorf("parse route path %q: %w", trimmed, err)
	}

	pathValue := frameworki18n.NormalizePath(parsed.Path)
	query := parsed.Query()
	removeInternalQueryMarkers(query)
	return pathValue, query, nil
}

func removeInternalQueryMarkers(query url.Values) {
	for key := range query {
		trimmed := strings.TrimSpace(key)
		if strings.HasPrefix(trimmed, "__") {
			query.Del(key)
		}
	}
}

func buildAbsoluteTypeURL(root *url.URL, href string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(href))
	if err != nil {
		return "", fmt.Errorf("parse href %q: %w", href, err)
	}

	if parsed.IsAbs() && strings.TrimSpace(parsed.Host) != "" {
		query := parsed.Query()
		removeInternalQueryMarkers(query)
		parsed.RawQuery = query.Encode()
		parsed.Fragment = ""
		return parsed.String(), nil
	}

	routePath := frameworki18n.NormalizePath(parsed.Path)
	query := parsed.Query()
	removeInternalQueryMarkers(query)
	return absoluteURL(root, routePath, query), nil
}

func absoluteURL(root *url.URL, routePath string, query url.Values) string {
	if root == nil {
		return routePath
	}

	clone := *root
	basePath := strings.TrimSuffix(strings.TrimSpace(clone.Path), "/")
	normalizedRoutePath := frameworki18n.NormalizePath(routePath)
	if normalizedRoutePath == "/" {
		if basePath == "" {
			clone.Path = "/"
		} else {
			clone.Path = basePath
		}
	} else {
		joined := path.Join(basePath, strings.TrimPrefix(normalizedRoutePath, "/"))
		if !strings.HasPrefix(joined, "/") {
			joined = "/" + joined
		}
		clone.Path = joined
	}

	if query == nil {
		clone.RawQuery = ""
	} else {
		clone.RawQuery = query.Encode()
	}
	clone.Fragment = ""
	return clone.String()
}
