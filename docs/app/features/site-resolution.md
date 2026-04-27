# Site Resolution

Use this guide when your app needs absolute canonical URLs for metadata, feeds,
sitemaps, or localized links.

## What This Feature Does

A `Site Resolver` gives `no-js` the absolute site root for features that need
to leave the current request path as an absolute URL.

That includes:

- canonical URLs
- alternate language URLs
- feed links
- sitemap entries
- request-scoped i18n URL helpers

## Pick The Right Resolver Shape

`framework/site` exposes two levels:

- `site.Resolver`
  String-based canonical and resolved URL methods.
- `site.URLResolver`
  Parsed `*url.URL` methods for safer URL composition.

The URL-aware form is the better default for new code.

## Static Resolver Example

```go
type StaticResolver struct {
	root *url.URL
}

func (r StaticResolver) CanonicalURL() string {
	return r.root.String()
}

func (r StaticResolver) Resolve(*http.Request) string {
	return r.root.String()
}

func (r StaticResolver) CanonicalRoot() *url.URL {
	clone := *r.root
	return &clone
}

func (r StaticResolver) ResolveRoot(*http.Request) *url.URL {
	return r.CanonicalRoot()
}
```

## Request-Aware Resolver Example

```go
type RequestResolver struct {
	root *url.URL
}

func (r RequestResolver) CanonicalURL() string {
	return r.root.String()
}

func (r RequestResolver) Resolve(*http.Request) string {
	return r.CanonicalURL()
}

func (r RequestResolver) CanonicalRoot() *url.URL {
	clone := *r.root
	return &clone
}

func (r RequestResolver) ResolveRoot(req *http.Request) *url.URL {
	root := r.CanonicalRoot()
	if root == nil || req == nil {
		return root
	}

	if req.Host != "" {
		root.Host = req.Host
	}
	if req.TLS != nil {
		root.Scheme = "https"
	}

	return root
}
```

## Where It Belongs

`Site Resolver` is an app-owned dependency. Put it in your runtime wiring, not
in `Custom Config`.

Keep site-root policy centralized there, whether the policy is static,
proxy-aware, or tenant-aware.

## Related Docs

- [Metadata and Head](metadata-and-head.md)
- [Discovery](discovery.md)
- [i18n](i18n.md)
