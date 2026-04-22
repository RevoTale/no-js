# Discovery

Use this guide when your app needs `robots.txt`, RSS feeds, or XML sitemaps.

## What This Feature Does

Discovery conventions let your app return structured data while the framework
owns routing, serialization, and cache headers.

Reserved files under `web/routes`:

- `robots.go`
- `feed.go`
- `sitemap.go`

`feed.go` and `sitemap.go` may live at the route root or inside nested
non-slot route directories.

## File To Endpoint Mapping

- `web/routes/robots.go` -> `/robots.txt`
- `web/routes/feed.go` -> `/feed.xml`
- `web/routes/sitemap.go` -> `/sitemap.xml` and `/sitemap-index.xml`
- `web/routes/author/_param__slug/feed.go` -> `/author/:slug/feed.xml`
- `web/routes/notes/sitemap.go` -> `/notes/sitemap.xml` and nested sitemap
  index endpoints

## Feed Example

```go
func Feed(
	runtime framework.RuntimeContext[*runtime.Context],
	r *http.Request,
) (discovery.FeedDocument, error) {
	return discovery.FeedDocument{
		Title:       "Notes",
		Link:        "https://example.com/notes",
		Description: "Latest notes",
	}, nil
}
```

Placed at `web/routes/feed.go`, this serves `/feed.xml`.

## Sitemap Example

Static sitemaps return `[]discovery.SitemapEntry`:

```go
func Sitemap(
	runtime framework.RuntimeContext[*runtime.Context],
	r *http.Request,
) ([]discovery.SitemapEntry, error) {
	return []discovery.SitemapEntry{
		{URL: "https://example.com/"},
		{URL: "https://example.com/notes"},
	}, nil
}
```

Large sitemaps can opt into chunk generation:

```go
func GenerateSitemaps(runtime framework.RuntimeContext[*runtime.Context], r *http.Request) ([]discovery.SitemapID, error)
func SitemapChunk(runtime framework.RuntimeContext[*runtime.Context], r *http.Request, id string) ([]discovery.SitemapEntry, error)
```

The framework serves chunk requests at `/sitemap/[id].xml` under the matched
route.

## `robots.txt`

`robots.go` returns `discovery.Robots`. The framework renders the text output.

Use it when you want the app to own allow/disallow rules but keep the transport
and serialization in one place.

## Caveat For Dynamic Nested Discovery

Nested discovery callbacks are still request-based. If the route is dynamic, the
callback reads what it needs from the request path or from app-owned helpers.

## Related Docs

- [Site Resolution](site-resolution.md)
- [HTTP Server and Runtime](httpserver-and-runtime.md)
- [Troubleshooting](../troubleshooting.md)
