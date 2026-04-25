# Static Assets And Client Assets

Use this guide when your app needs hashed asset URLs, colocated CSS/JS, or
fixed-path public files.

## What This Feature Does

`no-js` fingerprints bundled assets, writes a manifest, and serves the built
output under a versioned runtime prefix:

```text
/_assets/<hash>/
```

There are three asset paths:

- `web/routes` and `web/components` for colocated Client Assets.
- `web/assets` for explicit global files that should be fingerprinted.
- `web/public` for fixed request paths such as `/favicon.ico`.

## Client Assets

Colocated `.css`, `.js`, `.ts`, `.mjs`, and `.mts` files under route or
component packages are discovered automatically.

`no-js gen routes` writes source-adjacent Go helpers:

```text
web/routes/page.css       -> web/routes/page.css_gen.go
web/components/meter.ts   -> web/components/meter.ts_gen.go
```

CSS class constants are exported by default:

```go
const (
	PageShellClass = "n_a1b2c3d4"
)
```

Use those constants from templ:

```templ
<main class={ PageShellClass }>
```

The original class names stay in source CSS. The rendered HTML and built CSS use
anonymized class names.

## Route Bundles

`no-js gen assets` bundles matched route dependencies into route-level assets:

```html
<link rel="stylesheet" href="/_assets/<hash>/routes/index.css">
<script type="module" src="/_assets/<hash>/routes/index.js"></script>
```

Route assets are static per route. If a page imports a component package, that
component package's colocated CSS and scripts are included in that page's route
bundle. Pages that do not import the component do not receive its script.

## Script Helpers

For manual use, script source files also get exported helpers:

```go
func MeterScript() templ.Component
```

Normal pages should not call these helpers; route generation auto-injects route
scripts once per matched page. The helpers exist for Advanced composition cases
where you intentionally render a script outside the generated route flow.

TypeScript is bundled, not typechecked. If your app uses TypeScript, keep
`tsc --noEmit` in your app validation flow.

## Global Assets

Put explicit global files in `web/assets`, then run:

```bash
go tool no-js gen assets -root .
```

By default this writes processed assets and the manifest to `web/assets-build`.
At runtime, `httpserver.NewApp(...)` reads the manifest and serves the versioned
asset tree.

If you need to reference a global asset manually, prefer app-owned helpers. The
optional `view.SetStaticAssetBasePath(prefix string)` hook is still supported,
but it is not required for generated Client Assets.

## Legacy `-templ-css`

`-templ-css` remains available for templ `css` components:

```bash
go tool no-js gen assets -root . -templ-css
```

That generates `styles/templ.css` and sends it through the same hashed asset
pipeline. New apps should prefer colocated `.css` files when they want exported
class constants and route-level bundles.

If you need parameterized templ CSS variants, return them from
`TemplCSSVariants()` in `web/view`. If you only use zero-argument templ `css`
components, omit the hook.

## `web/assets` Versus `web/public`

Use `web/assets` for files that should be fingerprinted and versioned:

- `web/assets/site.css` -> `/_assets/<hash>/site.css`

Use `web/public` for fixed request paths:

- `web/public/favicon.ico` -> `/favicon.ico`
- `web/public/site.webmanifest` -> `/site.webmanifest`

Do not mix the two concerns.

## Runtime Overrides

If you need to override the manifest path or base URL prefix, do it in
`Custom Config`:

```go
Custom: httpserver.CustomConfig{
	StaticAssets: &httpserver.StaticAssetsConfig{
		ManifestPath: "web/assets-build/manifest.json",
		URLPrefix:    "/_assets/",
	},
}
```

Most apps should keep the default path.

## Related Docs

- [HTTP Server Reference](../reference/httpserver.md)
- [Bundle Config Reference](../reference/bundle-config.md)
- [Metadata and Head](metadata-and-head.md)
