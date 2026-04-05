# Request Cache and Partials

## What This Feature Does

`no-js` gives page loaders a request-scoped cache and handles HTMX partial
requests without changing your page loader contract.

The request cache deduplicates repeated work inside one request. Partial request
handling lets the framework skip the root layout and send metadata patches when a
page is refreshed through HTMX.

## Modules

- `framework` (`request_cache.go`)
- `framework/metagen`
- `framework/httpserver`

## Happy Path

Every page request is wrapped with `framework.WithRequestCache(...)` before
loaders run. Use `framework.CachedCall(...)` inside loaders to share results
within the same request.

## Focused Example

```go
func LoadAuthorPage(
	ctx context.Context,
	appCtx *runtime.Context,
	r *http.Request,
	params framework.SlugParams,
) (runtime.AuthorPageView, error) {
	cacheKey := "LoadAuthorPage:" + params.Slug

	return framework.CachedCall(ctx, cacheKey, func(runCtx context.Context) (runtime.AuthorPageView, error) {
		return loadAuthorPage(runCtx, appCtx, r, params.Slug)
	})
}
```

The cache key is app-defined. The sharing scope is one request only.

## HTMX Partials

When the request includes `HX-Request`, the framework:

- renders the page component without the root layout
- keeps running the same metadata resolver
- sends the metadata patch in HTMX response headers

That behavior is automatic. Most apps do not need to branch on partial requests
unless they want different caching or transport behavior.

## Related Docs

- [Metadata and Head](metadata-and-head.md)
- [HTTP Server and Runtime](httpserver-and-runtime.md)
