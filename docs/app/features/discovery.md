# Discovery

## What This Feature Does

Discovery conventions let your app return structured data for `robots.txt`,
RSS feeds, and XML sitemaps while the framework owns routing, serialization, and
cache headers.

Reserved files live under `web/routes`:

- `robots.go`
- `feed.go`
- `sitemap.go`

`feed.go` and `sitemap.go` may live at the route root or inside nested route
directories.

## Modules

- `framework/discovery`

## Happy Path

A root feed is just a function that returns `discovery.FeedDocument`:

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

## Focused Example

Nested discovery files inherit their route directory. For example:

```text
web/routes/author/_param__slug/feed.go
```

serves:

```text
/author/:slug/feed.xml
```

The current discovery callback is still request-based, so dynamic nested routes
read params from the request path or app-owned helpers:

```go
func Feed(
	runtime framework.RuntimeContext[*runtime.Context],
	r *http.Request,
) (discovery.FeedDocument, error) {
	slug := strings.TrimPrefix(r.URL.Path, "/author/")
	slug = strings.TrimSuffix(slug, "/feed.xml")

	return discovery.FeedDocument{
		Title: "Author feed",
		Link:  "https://example.com/author/" + slug,
	}, nil
}
```

That part is more manual than page routes. If you need shared parsing, keep it in
an app-owned helper.

## Sitemap Chunks

Static sitemaps return `[]discovery.SitemapEntry`.

Large sitemaps can opt into chunk generation with both of these functions:

```go
func GenerateSitemaps(runtime framework.RuntimeContext[*runtime.Context], r *http.Request) ([]discovery.SitemapID, error)
func SitemapByID(runtime framework.RuntimeContext[*runtime.Context], r *http.Request, id string) ([]discovery.SitemapEntry, error)
```

When both are present, the framework serves sitemap-index endpoints and chunk
URLs automatically.

## Related Docs

- [App Conventions](../conventions.md)
- [Site Resolution](site-resolution.md)
- [HTTP Server and Runtime](httpserver-and-runtime.md)
