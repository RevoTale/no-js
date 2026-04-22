# Metadata And Head

Use this guide when you are composing titles, canonical URLs, alternates, or
Open Graph/Twitter metadata.

## What This Feature Does

`no-js` separates metadata composition from page rendering.

Your metadata resolvers receive a `MetaContext[C]` with:

- the current request
- the current locale
- the resolved site root
- URL helpers
- access to the app context

Your root template renders the result through `@metagen.Head(meta)`.

## Render The Managed Head In The Root Layout

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

That output includes:

- title
- description
- canonical and alternate links
- stylesheet links managed by the framework
- robots directives
- Open Graph and Twitter tags

## Use `MetaContext` Helpers

Prefer `MetaContext` helpers instead of hand-building URLs:

```go
func (Resolver) MetaGenAuthorParamSlugPage(
	meta framework.MetaContext[*runtime.Context],
	params AuthorParamSlugParams,
) (metagen.Metadata, error) {
	canonical := meta.LocalizedURL(meta.Locale(), "/author/"+params.Slug)
	alternates, err := meta.Alternates(meta.Locale(), nil)
	if err != nil {
		return metagen.Metadata{}, err
	}

	return metagen.Metadata{
		Title:      "Author",
		Alternates: alternates,
		OpenGraph: &metagen.OpenGraph{
			Type: "profile",
			URL:  canonical.String(),
		},
	}, nil
}
```

This keeps canonical URLs, localized URLs, and alternates on the same site-root
policy.

## Managed Stylesheets

`@metagen.Head(meta)` is also where framework-managed stylesheet links appear.

That includes:

- hashed asset stylesheets that you add to `metagen.Metadata.Stylesheets`
- build-time templ CSS output when you use `-templ-css`

## HTMX Partial Requests

On HTMX partial requests, `no-js`:

- renders the page body without the root layout
- keeps running the same metadata resolver
- sends a metadata patch through response headers

Most apps keep one metadata path for both full-page and HTMX rendering.

## Trusted Escape Hatch

If you need to inject trusted raw head HTML, `metagen.Metadata` also exposes
`DangerRawHead`. Do not populate it from user-controlled or third-party input.

## Related Docs

- [Site Resolution](site-resolution.md)
- [i18n](i18n.md)
- [Request Cache and Partials](request-cache-and-partials.md)
