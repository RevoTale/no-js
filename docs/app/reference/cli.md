# CLI Reference

`no-js` is the public build-time entrypoint for consuming apps.

## Synopsis

```bash
go tool no-js gen [routes|assets|check] [-root .] [-config path] [-templ-css]
```

If you omit the mode, `gen` runs both route generation and asset generation.

## Modes

- `go tool no-js gen -root .`
  Run the normal generation loop.
- `go tool no-js gen routes -root .`
  Generate route handlers, resolver contracts, built-in i18n output, and the
  templ CSS registry.
- `go tool no-js gen assets -root .`
  Build the static asset bundle and write the manifest.
- `go tool no-js gen check -root .`
  Run route generation, asset generation, then fail if `git diff --exit-code`
  is not clean.

## Flags

- `-root`
  Application root directory. Default: `.`
- `-config`
  Explicit bundle-config path. If omitted, `no-js` loads
  `no-js.bundle.yaml` from the app root when the file exists.
- `-templ-css`
  Generate `styles/templ.css` from templ `css` components before asset
  bundling.

## What Each Run Produces

`go tool no-js gen routes -root .` writes:

- `web/generated/*`
- `web/resolvers/generated.go`
- built-in i18n output when `web/i18n/messages` exists
- templ CSS registry output used by the runtime and optional asset bundling

`go tool no-js gen assets -root .` writes:

- processed files under the configured assets-build directory
- the static asset manifest, by default `web/assets-build/manifest.json`

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

Bundle templ CSS into the hashed asset pipeline:

```bash
go tool no-js gen assets -root . -templ-css
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
- [Bundle Config Reference](bundle-config.md)
- [Static Assets](../features/static-assets.md)
