# Request Cache And Partials

Use this guide when you want to reuse loader work inside one request or support
HTMX partial refreshes without changing your loader contract.

## What This Feature Does

`no-js` gives page loaders a request-scoped cache and handles HTMX partial
requests automatically.

That means:

- repeated work inside one request can be deduplicated
- HTMX partial requests can reuse the same loader and metadata path
- the framework decides when to skip the root layout and emit metadata patches

## Request Cache

Every page request is wrapped with `framework.WithRequestCache(...)` before
loaders run. Use `framework.CachedCall(...)` inside a loader to share work:

```go
func (Resolver) ResolveAuthorParamSlugPage(
	ctx context.Context,
	appCtx *runtime.Context,
	r *http.Request,
	params AuthorParamSlugParams,
) (runtime.AuthorPageView, error) {
	cacheKey := "author:" + params.Slug

	shared, err := framework.CachedCall(ctx, cacheKey, func(runCtx context.Context) (string, error) {
		return loadAuthor(runCtx, appCtx, params.Slug)
	})
	if err != nil {
		return runtime.AuthorPageView{}, err
	}

	return runtime.AuthorPageView{
		Heading: shared,
	}, nil
}
```

The cache key is app-defined. The sharing scope is one request only.

## HTMX Partial Requests

When the request includes `HX-Request`, the framework:

- renders the page component without the root layout
- keeps using the same metadata resolver
- sends the metadata patch in response headers

Most apps do not need special branching for partial requests unless they want
different caching or transport behavior.

## What Stays The Same

- the generated route contract
- page loaders
- metadata resolvers
- app context usage

The framework changes the render behavior, not the app-facing route API.

## Related Docs

- [Metadata and Head](metadata-and-head.md)
- [HTTP Server and Runtime](httpserver-and-runtime.md)
