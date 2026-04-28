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
- module script links managed by the framework
- robots directives
- Open Graph and Twitter tags

## Use `MetaContext` Helpers

Prefer `MetaContext` helpers instead of hand-building URLs:

```go
func (Resolver) MetaGenAuthorParamSlugPage(
	meta framework.MetaContext[*view.Context],
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

## 404 Metadata

Generated 404 rendering uses `MetaGen...NotFound(...)` methods for not-found
head fields. The metadata chain is root layout, matched route layouts, then the
matched 404 metadata resolver. Generated rendering still applies
`noindex, nofollow` to 404 responses.

```go
func (Resolver) MetaGenRootNotFound(
	meta framework.MetaContext[*view.Context],
	notFound framework.NotFoundContext,
	params RootParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Not Found"}, nil
}
```

## Managed Client Assets

`@metagen.Head(meta)` is also where framework-managed Client Assets appear.

That includes:

- shell, layout-subtree, and page-fallback CSS generated from colocated `.css` files
- owner module scripts generated from colocated `.js`, `.ts`, `.tsx`, `.mjs`, and `.mts` files
- hashed asset stylesheets that you add to `metagen.Metadata.Stylesheets`
- the global `styles/templ.css` stylesheet unless `assets.templ_css` is false

Managed stylesheets render before managed module scripts. Route generation
resolves their hashed URLs at runtime, so app templates do not need global asset
base-path state for generated Client Assets.

Default templ CSS extraction is separate from Client Asset CSS. It collects
registered templ `css {}` classes into one global stylesheet and injects that
stylesheet on every page through the same managed head path. Set
`assets.templ_css` to `false` to disable it.

Files under `web/assets` are not auto-injected. Add those URLs to metadata
yourself when you intentionally want a global hashed asset on a page:

```go
return metagen.Metadata{
	Stylesheets: []string{
		metagen.AssetURL(meta.Context(), "site.css"),
	},
}, nil
```

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
