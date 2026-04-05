# Metadata and Head

## What This Feature Does

`no-js` separates metadata composition from page rendering. Route metadata
resolvers receive a `MetaContext[C]` with request-scoped URL helpers, locale,
root resolution, and access to the app context.

Your root template renders the result through `@metagen.Head(meta)`.

## Modules

- `framework/metadata_context`
- `framework/metagen`

## Happy Path

Metadata resolvers use `MetaContext[*runtime.Context]`:

```go
func (Resolver) MetaGenNotePage(
	meta framework.MetaContext[*runtime.Context],
	params framework.SlugParams,
) (metagen.Metadata, error) {
	canonical := meta.LocalizedURL(meta.Locale(), "/note/"+params.Slug)
	alternates, err := meta.Alternates(meta.Locale(), map[string]string{
		"application/rss+xml": "/feed.xml",
	})
	if err != nil {
		return metagen.Metadata{}, err
	}

	return metagen.Metadata{
		Title: "Note",
		Alternates: alternates,
		OpenGraph: &metagen.OpenGraph{
			URL: canonical.String(),
		},
	}, nil
}
```

The root template inserts the managed head:

```templ
templ RootLayout(meta metagen.Metadata, locale string, child templ.Component) {
	<html lang={ locale }>
		<head>
			@metagen.Head(meta)
		</head>
		<body>@child</body>
	</html>
}
```

## Focused Example

Use `MetaContext` helpers instead of hand-joining URLs:

```go
func MetaGenAuthorPage(
	meta framework.MetaContext[*runtime.Context],
	slug string,
) (metagen.Metadata, error) {
	profileURL := meta.LocalizedURL(meta.Locale(), "/author/"+slug)
	alternates, err := meta.Alternates(meta.Locale(), nil)
	if err != nil {
		return metagen.Metadata{}, err
	}

	return metagen.Metadata{
		Alternates: alternates,
		OpenGraph: &metagen.OpenGraph{
			Type: "profile",
			URL:  profileURL.String(),
		},
	}, nil
}
```

This keeps canonical URLs, hreflang alternates, and request-aware site roots on
the same path.

## HTMX Partials

On HTMX partial requests, `no-js` renders the page body without the root layout
and sends a metadata patch through response headers. You keep the same metadata
resolver. The framework owns the patch format and header writing.

## Related Docs

- [Routing and Generation](routing-and-generation.md)
- [Site Resolution](site-resolution.md)
- [Request Cache and Partials](request-cache-and-partials.md)
