# Static Assets

Use this guide when your app needs hashed asset URLs, fixed-path public files,
or templ CSS bundled into the asset pipeline.

## What This Feature Does

`no-js` fingerprints bundled assets, writes a manifest, and serves the built
output under a versioned runtime prefix.

This is separate from fixed-path public files under `web/public`.

## Happy Path

Put source assets in `web/assets`, then run:

```bash
go tool no-js gen assets -root .
```

By default this writes:

- processed assets to `web/assets-build`
- the manifest to `web/assets-build/manifest.json`

At runtime, `httpserver.NewApp(...)` reads that manifest and serves assets
under:

```text
/_assets/<hash>/
```

## Template Usage

Most apps expose a small helper in `web/view`:

```go
package view

import (
	"path"
	"strings"
)

var staticAssetBasePath string

func SetStaticAssetBasePath(prefix string) {
	staticAssetBasePath = strings.TrimRight(strings.TrimSpace(prefix), "/")
}

func StaticAssetURL(assetPath string) string {
	trimmed := strings.TrimPrefix(strings.TrimSpace(assetPath), "/")
	if trimmed == "" {
		return staticAssetBasePath
	}
	return path.Join(staticAssetBasePath, trimmed)
}
```

Then templates use that helper:

```templ
<link rel="stylesheet" href={ view.StaticAssetURL("site.css") }>
```

`generated.Bundle(appContext)` wires `SetStaticAssetBasePath(...)` only when
your `web/view` package defines it. If you use a different asset URL helper,
you can skip that hook entirely.

## `-templ-css`

If you want templ `css` components to become a hashed static stylesheet, add
`-templ-css`:

```bash
go tool no-js gen assets -root . -templ-css
```

That generates `styles/templ.css` and sends it through the same hashed asset
pipeline as the rest of `web/assets`.

Without `-templ-css`, templ `css` output stays on templ's normal render path and
does not go through the static asset manifest.

## Zero-Arg CSS And Explicit Variants

The generated CSS registry automatically picks up zero-argument templ `css`
components from:

- `web/routes`
- `web/components`

If you need parameterized variants, return them from `TemplCSSVariants()` in
`web/view`. For example, if `web/components` exposes
`ProgressBar(percent int)`:

```go
package view

import (
	"example.com/your-app/web/components"
	"github.com/a-h/templ"
)

func TemplCSSVariants() []templ.CSSClass {
	return []templ.CSSClass{
		components.ProgressBar(72),
	}
}
```

If you only use zero-argument `css` components, omit the hook.

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
