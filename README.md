# no-js

[![Status: Work in Progress](https://img.shields.io/badge/status-work%20in%20progress-orange)](#)

`no-js` is an opinionated Go framework for server-rendered web applications.

## What You Build With It

`no-js` is for apps that want:

- server-rendered pages with typed loaders and layouts
- generated route wiring from a strict file-based app tree
- metadata composition for `<head>`
- i18n-aware routing
- static asset fingerprinting
- optimized data streaming 
- optional HTMX partial navigation support

## Required App Structure

The route generator is strict. A consuming app is expected to look like this:

```text
your-app/                             # provides the app root; required
  go.mod                              # provides the module path for generated imports; required
  internal/                           # provides the framework-visible app namespace; required
    web/                              # provides the fixed web namespace used by the generator; required
      app/                            # provides the file-based route tree; required
        root.templ                    # provides the document shell; required
        404.templ                     # provides the root not-found page; required
        error.templ                   # provides the root error page; required
        page.templ                    # provides the / route page; optional unless you serve /
        layout.templ                  # provides the root-segment layout; optional
        note/                         # provides a static route segment; optional
          [slug]/                     # provides a dynamic route segment; optional, but [param] syntax is required
            page.templ                # provides the /note/:slug page; required for this route
      runtime/                        # provides app-specific contracts for generated code; required today
        context.go                    # provides runtime.Context; required today
        view_models.go                # provides runtime page view types and RootLayoutView; required today
        loaders.go                    # provides page loaders and metadata helpers; usually required
      bootstrap/                      # provides app-owned startup composition for generated server config; recommended
      resolvers/                      # provides handwritten route resolver methods; required
        root.go                       # provides resolver methods for the root route; optional per route
        note_param_slug.go            # provides resolver methods for /note/[slug]; optional per route
```

Generated files are written to:

```text
internal/web/gen/                     # provides generated route modules with safe Go package names
internal/web/resolvers/generated.go   # provides generated route resolver interfaces and param types
```

## Conventions

These are not suggestions. They are current framework contracts.

- Routes live under `internal/web/app`.
- Dynamic segments must use `[param]` directories.
- Route templates must use the exact file names `root.templ`, `layout.templ`, `page.templ`, `404.templ`, and `error.templ`.
- `root.templ` is required at `internal/web/app/root.templ`.
- Root `404.templ` and root `error.templ` are required.
- `page.templ` view types must be `runtime.*`.
- Layout and error/not-found contracts currently depend on `runtime.RootLayoutView`.
- Generated code imports `internal/web/runtime` and `internal/web/resolvers` from the consuming module.
- App-owned startup wiring is expected to live under `internal/web/bootstrap` by default.
- Route-local `components/` directories are rejected by the generator.
- Only `root.templ` may contain document-level tags like `<html>`, `<head>`, and `<body>`.

## Build Config

`no-js` supports an optional root config file named `no-js.bundle.yaml`.

This file is for build-time configuration only. It controls deterministic inputs such as project layout paths,
generated server feature flags, static asset build settings, public file serving defaults, and the path to the
app-owned bootstrap package.

If `no-js.bundle.yaml` is missing, the CLI uses framework defaults. A missing config file is not an error.

If the file exists, `version: 1` is required and unknown fields are rejected.

YAML values override defaults field by field. Unspecified fields keep their default values.

There are no globally required operational YAML fields. Requirements are resolved from the command and enabled
features. For example, `go.mod` and the resolved app/runtime directories must exist, the resolved i18n directory must
exist if i18n routing is enabled, static asset paths must be valid when running static asset generation, and the
public directory must exist only when public file support is enabled and used. The configured `bootstrap_dir` is
resolved as part of project layout, but bundling does not import or execute it.

Invalid YAML is an error.

`no-js.bundle.yaml` must not contain runtime or environment-specific values. Keep process-time configuration such as
listen address, root URL, API tokens, analytics IDs, cache overrides, static asset URL prefix, and service wiring in
app-owned Go bootstrap code.

Example default-equivalent config:

```yaml
version: 1

project:
  app_dir: internal/web/app
  gen_dir: internal/web/gen
  resolver_dir: internal/web/resolvers
  runtime_dir: internal/web/runtime
  bootstrap_dir: internal/web/bootstrap
  i18n_dir: internal/web/i18n
  public_dir: public

server:
  features:
    i18n_routing: auto
    static_assets: auto
    public_files: auto
    health_endpoint: auto

static_assets:
  source_dir: internal/web/static
  out_dir: internal/web/static-build
  manifest_path: internal/web/static-build/manifest.json

public_files:
  request_path_prefix: /
```

`manifest_path` points to generated static-bundle metadata. It stores the asset hash used by the runtime to construct
the final versioned asset prefix. The public static asset URL prefix is runtime config, not YAML config.

## What We Support

- Runtime packages under `framework/*`
- `framework/engine`
  Route execution, concurrent metadata and page loading, layout composition, and streaming root-layout rendering.
- `framework/httpserver`
  HTTP server integration, cache policies, gzip, `/healthz`, static asset mounting, and optional public-file middleware.
- `framework/metagen`
  Canonical URLs, alternate languages/types, robots tags, Open Graph, Twitter cards, Pinterest tags, and HTMX head patch generation.
- `framework/i18n`
  Locale config, locale-aware path handling, request locale context, and routing prefix modes: `always`, `as-needed`, `never`.
- `framework/staticassets`
  Runtime manifest loading and runtime asset URL composition.

- Build-time packages under `bundler/*`
- `bundler/approutegen`
  Route discovery and generated registry/resolver contracts from the file tree.
- `bundler/i18nkeygen`
  Go key generation from canonical locale message definitions.
- `bundler/staticassets`
  Minification, hashing, manifest generation, and versioned asset bundle assembly.
- `bundler/templgen`
  `templ` generation for selected files or paths.

- CLI entrypoints under `cmd/*`
- `cmd/no-js`
  Recommended root CLI for `gen`, `gen routes`, `gen assets`, and `gen check`.
- `cmd/*`
  Low-level CLI entrypoints for route generation, `templ` generation, i18n key generation, and static asset building.

## Nuances

- This framework is intentionally not generic yet. The generator still assumes an `runtime` package and specific template signatures.
- The generator is module-aware: framework imports point to `github.com/RevoTale/no-js`, but generated app imports are resolved from the consuming app's `go.mod`.
- i18n locales are currently normalized to two-letter lowercase codes.
- HTMX support is request-driven. Partial requests are detected through `HX-Request`, and metadata patches are emitted through response headers.
- Build-time packages may import runtime-owned shared types, but runtime packages must not import bundler packages.
- Static assets and public files are separate concerns:
  `/_assets/` is the default runtime prefix for fingerprinted build output, while public files are served as fixed
  request paths.

## Development

```bash
task fix
task validate
task test
```

## Origin

`no-js` originated as an extraction from [RevoTale/blog](https://github.com/RevoTale/blog).
