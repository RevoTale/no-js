# CLI Reference

`no-js` is the public build-time entrypoint for consuming apps.

The supported app-facing tools are:

- `go tool no-js`
  Main generation entrypoint for routes, assets, checks, and bundle-config
  loading.
- `go tool templgen`
  Companion tool only for extra `.templ` packages outside `web/routes`.

Do not add lower-level repository generators such as `approutegen`,
`staticassetsgen`, `templcssgen`, or `i18nkeygen` to consuming app toolchains.
They are implementation helpers for this repository, not the app CLI contract.

## Synopsis

```bash
go tool no-js gen [routes|assets|check] [-root .] [-config path]
```

If you omit the mode, `gen` runs both route generation and asset generation.
Every mode first validates the strict `web/routes` and `web/components` app
shape. Think of `routes` as the generation step that writes Go code, and
`assets` as the asset step that writes bundled, fingerprinted runtime files.

## Modes

- `go tool no-js gen -root .`
  Run the normal generation loop.
- `go tool no-js gen routes -root .`
  Generate Go route handlers, resolver contracts, built-in i18n output,
  source-adjacent Client Asset helpers, and the templ CSS registry. This does
  not write final browser CSS or JavaScript bundles.
- `go tool no-js gen assets -root .`
  Build route-level Client Asset bundles, include explicit global files from
  `web/assets`, fingerprint the output, and write the manifest.
- `go tool no-js gen check -root .`
  Run route generation, asset generation, then fail if `git diff --exit-code`
  is not clean.

## Flags

- `-root`
  Application root directory. Default: `.`
- `-config`
  Explicit bundle-config path. If omitted, `no-js` loads
  `no-js.bundle.yaml` from the app root when the file exists.

## What Each Run Produces

`go tool no-js gen routes -root .` writes:

- `web/generated/*`
- `web/resolvers/generated.go`
- built-in i18n output when `web/i18n/messages` exists
- `*.css_gen.go` and `*.{js,ts,tsx,mjs,mts}_gen.go` beside discovered Client Asset sources
- templ CSS registry output unless `assets.templ_css` is false

`go tool no-js gen assets -root .` writes:

- generated Client Asset stylesheets, module entries, and shared script chunks
- explicit global `web/assets` files under the configured assets-build directory
- the static asset manifest, by default `web/assets-build/manifest.json`

Client Asset scripts are bundled with esbuild and can use JavaScript or
TypeScript imports. `web/assets` files are path-managed: CSS may bundle relative
CSS imports, while JavaScript is minified as a standalone file. All esbuild
paths use `assets.browser_targets` from bundle config.

See [Asset Pipeline Reference](asset-pipeline.md) for the compile, bundle, and
fingerprint split.

## Common Examples

Generate everything from the app root:

```bash
go tool no-js gen -root .
```

Check generated output in CI:

```bash
go tool no-js gen check -root .
```

Use an explicit config file:

```bash
go tool no-js gen -root . -config ./config/no-js.bundle.yaml
```

Disable templ `css` extraction when you want templ to keep inline CSS registration:

```yaml
version: 1

assets:
  templ_css: false
```

```bash
go tool no-js gen assets -root .
```

## When You Also Use `templgen`

`go tool no-js` handles `web/routes` and the generated route output. If your app
has extra `.templ` files under `web/components`, `web/view`, or other
app-owned packages, use the companion `templgen` tool as a second step.

Add it to `go.mod`:

```go
tool (
	github.com/RevoTale/no-js/cmd/no-js
	github.com/RevoTale/no-js/cmd/templgen
)
```

Run it:

```bash
go tool templgen -base . -path web/components -path web/view -path web/generated
```

Supported `templgen` flags:

- `-base`
  Base path used for relative filenames in generated output. Default: `.`
- `-path`
  Directory to scan for `.templ` files. Repeatable.
- `-file`
  Single `.templ` file to compile. Repeatable.

Keep `go tool no-js` as the main entrypoint. Add `templgen` only for extra
templ packages outside `web/routes`.

## Related Docs

- [Getting Started](../getting-started.md)
- [Asset Pipeline Reference](asset-pipeline.md)
- [Bundle Config Reference](bundle-config.md)
- [Static Assets](../features/static-assets.md)
