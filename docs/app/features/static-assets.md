# Static Assets

## What This Feature Does

`no-js` fingerprints bundled assets, writes a manifest, and serves the built
output under a versioned runtime prefix. This is separate from fixed-path public
files under `web/public`.

## Modules

- `framework/staticassets`
- `framework/httpserver`
- `internal/bundler/staticassets`

## Happy Path

Asset generation reads from `web/assets` and writes to `web/assets-build`:

```bash
go run github.com/RevoTale/no-js/cmd/no-js gen assets -root .
```

To build a global templ stylesheet and send it through the same hashed asset
pipeline:

```bash
go run github.com/RevoTale/no-js/cmd/no-js gen assets -root . -templ-css
```

The generated manifest lives at:

```text
web/assets-build/manifest.json
```

At runtime, `httpserver.NewApp(...)` reads that manifest and serves assets under
the default versioned prefix:

```text
/_assets/<hash>/
```

## Focused Example

If you need to override the defaults, do it in `Custom Config`:

```go
handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
	App: generated.Bundle(appContext),
	Custom: httpserver.CustomConfig{
		StaticAssets: &httpserver.StaticAssetsConfig{
			ManifestPath: "web/assets-build/manifest.json",
			URLPrefix:    "/_assets/",
		},
	},
})
```

At the app layer, templates usually point to the resolved runtime prefix through
a small helper:

```templ
<link rel="stylesheet" href={ runtime.StaticAssetURL("app.css") }/>
```

For templ `css` components, opt in at runtime so registered classes are treated
as globally available:

```go
bundle := generated.Bundle(appContext)
bundle.TemplCSSClasses = generated.TemplCSSClasses
```

`generated.TemplCSSClasses()` auto-registers zero-arg templ `css` components
from `web/routes` and `web/components`. Explicit variants still belong in
`web/view.TemplCSSVariants() []templ.CSSClass`.

## `/_assets/` Versus `web/public`

Use `web/assets` for files that should be fingerprinted and served from the
versioned prefix.

Use `web/public` for fixed request paths:

- `web/assets/app.css` -> `/_assets/<hash>/app.css`
- `web/public/favicon.ico` -> `/favicon.ico`
- `web/public/site.webmanifest` -> `/site.webmanifest`

Do not mix the two concerns.

## When To Use `Custom Config` Or Advanced Composition

Stay on the default manifest and prefix unless you have a deployment constraint
that forces an override.

If your app needs a different template helper or CDN policy, keep that in
app-owned runtime code. The manifest and static mount stay framework-owned.

## Related Docs

- [Getting Started](../getting-started.md)
- [HTTP Server and Runtime](httpserver-and-runtime.md)
