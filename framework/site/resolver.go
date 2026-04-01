package site

import (
	"net/http"
	"net/url"
	"strings"
)

type Resolver interface {
	CanonicalURL() string
	Resolve(*http.Request) string
}

type URLResolver interface {
	Resolver
	CanonicalRoot() *url.URL
	ResolveRoot(*http.Request) *url.URL
}

func ResolveRootURL(resolver Resolver, r *http.Request) string {
	if resolver == nil {
		return ""
	}

	if r != nil {
		if resolved := strings.TrimSpace(resolver.Resolve(r)); resolved != "" {
			return resolved
		}
	}

	return strings.TrimSpace(resolver.CanonicalURL())
}

func ResolveRoot(resolver Resolver, r *http.Request) *url.URL {
	if resolver == nil {
		return nil
	}

	if urlResolver, ok := resolver.(URLResolver); ok {
		if r != nil {
			if resolved := normalizeRootURL(urlResolver.ResolveRoot(r)); resolved != nil {
				return resolved
			}
		}
		return normalizeRootURL(urlResolver.CanonicalRoot())
	}

	return parseRootURL(ResolveRootURL(resolver, r))
}

func parseRootURL(value string) *url.URL {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	return normalizeRootURL(parsed)
}

func normalizeRootURL(root *url.URL) *url.URL {
	if root == nil {
		return nil
	}

	clone := *root
	if !clone.IsAbs() || strings.TrimSpace(clone.Host) == "" {
		return nil
	}
	clone.RawQuery = ""
	clone.Fragment = ""
	return &clone
}
