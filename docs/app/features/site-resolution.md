# Site Resolution

## What This Feature Does

A `Site Resolver` gives `no-js` the absolute site root for canonical URLs,
alternates, feeds, sitemaps, and other metadata that must leave the current page
as an absolute URL.

If your app only has one canonical root, use a static resolver. If the public
host depends on the request, implement the URL-aware resolver methods.

## Modules

- `framework/site`
- `framework/metadata_context`

## Happy Path

`framework/site` supports two levels:

- `site.Resolver`
  string-based canonical and resolved URL methods
- `site.URLResolver`
  parsed `*url.URL` methods for safer URL composition

The URL-aware form is the better default for new code.

## Focused Example

Static resolver:

```go
type StaticResolver struct {
	root *url.URL
}

func (r StaticResolver) CanonicalURL() string          { return r.root.String() }
func (r StaticResolver) Resolve(*http.Request) string  { return r.root.String() }
func (r StaticResolver) CanonicalRoot() *url.URL       { clone := *r.root; return &clone }
func (r StaticResolver) ResolveRoot(*http.Request) *url.URL { return r.CanonicalRoot() }
```

Request-aware resolver:

```go
type RequestResolver struct {
	root *url.URL
}

func (r RequestResolver) CanonicalURL() string         { return r.root.String() }
func (r RequestResolver) Resolve(*http.Request) string { return r.CanonicalURL() }
func (r RequestResolver) CanonicalRoot() *url.URL      { clone := *r.root; return &clone }

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

Pass it into your app context once, then the same root flows into
`MetaContext.Root()`, `MetaContext.LocalizedURL(...)`, discovery builders, and
request-scoped i18n URL helpers.

## When To Use `Custom Config` Or Advanced Composition

`Site Resolver` is an app-owned dependency, not a `Custom Config` hook.

Put the resolver in app runtime wiring and keep its policy centralized there. If
you need tenant-aware or proxy-aware logic, that is still the same `Site Resolver`
contract, just with app-owned implementation.

## Related Docs

- [Metadata and Head](metadata-and-head.md)
- [Discovery](discovery.md)
- [i18n](i18n.md)
