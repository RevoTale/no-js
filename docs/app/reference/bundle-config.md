# Bundle Config Reference

`no-js.bundle.yaml` is optional. It controls build-time inputs only. Most apps
should omit it until they need to change a default; when you add it, include
only the keys you want to override.

If the file is missing, `no-js` uses defaults. If the file exists, it must
declare `version: 1` and unknown fields are rejected.

## Default Config

```yaml
version: 1

project:
  routes_dir: web/routes
  generated_dir: web/generated
  resolvers_dir: web/resolvers
  view_dir: web/view
  i18n_dir: web/i18n
  assets_dir: web/assets
  assets_build_dir: web/assets-build

server:
  features:
    i18n_routing: auto
    static_assets: auto
    health_endpoint: auto

i18n:
  mode: auto

assets:
  templ_css: true

static_assets:
  manifest_path: web/assets-build/manifest.json
```

## `project`

- `project.routes_dir`
  Route tree root. Default: `web/routes`
- `project.generated_dir`
  Generated route output directory. Default: `web/generated`
- `project.resolvers_dir`
  Handwritten resolver directory. Default: `web/resolvers`
- `project.view_dir`
  Handwritten view package. Default: `web/view`
- `project.i18n_dir`
  App i18n package root. Default: `web/i18n`
- `project.assets_dir`
  Source directory for explicit global hashed assets. Default: `web/assets`
- `project.assets_build_dir`
  Output directory for generated Client Asset bundles, explicit global assets,
  and the static manifest. Default: `web/assets-build`

All configured paths must stay relative to the app root.

## `server.features`

Supported values: `auto`, `enabled`, `disabled`

- `server.features.i18n_routing`
  Locale-prefixed routing support. `auto` enables it when `web/i18n` exists.
- `server.features.static_assets`
  Build explicit global hashed assets from `web/assets`. `auto` enables it
  when `web/assets` exists. You do not need this flag for colocated route or
  component CSS/JS/TS, or for templ CSS; those pipelines build their own assets
  when their source files exist. At startup, no-js serves whatever the build
  wrote to `web/assets-build`.
- `server.features.health_endpoint`
  Health endpoint support. `auto` enables it by default.

## `i18n`

Supported values: `auto`, `enabled`, `disabled`

- `i18n.mode`
  Built-in i18n code generation. `auto` enables it when
  `web/i18n/messages` exists.

If you want to keep the route/runtime primitives but own localization yourself,
set:

```yaml
version: 1

i18n:
  mode: disabled
```

## `assets`

- `assets.templ_css`
  Global templ CSS extraction. Default: `true`.

  By default, `no-js` scans templ `css {}` declarations and
  `web/view.TemplCSSVariants()`. When no zero-argument templ CSS declarations
  or variants hook exist, no `styles/templ.css` asset is emitted. Set this to
  `false` to keep templ CSS on templ's inline registration path.

```yaml
version: 1

assets:
  templ_css: false
```

## `static_assets`

- `static_assets.manifest_path`
  Manifest file path relative to the app root. Default:
  `web/assets-build/manifest.json`

This path controls where the asset pipeline writes the manifest that
`httpserver.NewApp(...)` later reads at runtime.

## Validation Rules

- `version: 1` is required
- unknown YAML fields are rejected
- configured paths must be relative and stay inside the app root
- `assets.templ_css` must be `true` or `false`
- `go.mod`, the configured route root, and the configured view root must exist
- if i18n routing is enabled, the configured i18n root must exist
- if built-in i18n is enabled, the configured messages directory must exist

## What Does Not Belong In This File

Do not put runtime or environment-specific values in `no-js.bundle.yaml`.

Keep these in app-owned Go code instead:

- listen addresses
- site or canonical URL policy
- middleware wiring
- service construction
- API tokens
- analytics IDs
- cache overrides tied to runtime behavior

## Example: Custom Paths

```yaml
version: 1

project:
  routes_dir: src/web/routes
  generated_dir: src/web/generated
  resolvers_dir: src/web/resolvers
  view_dir: src/web/view
  i18n_dir: src/web/i18n
  assets_dir: web-static
  assets_build_dir: web-static-build

static_assets:
  manifest_path: web-static-build/manifest.json
```

## Related Docs

- [CLI Reference](cli.md)
- [Asset Pipeline Reference](asset-pipeline.md)
- [App Conventions](../conventions.md)
- [i18n](../features/i18n.md)
- [Static Assets](../features/static-assets.md)
