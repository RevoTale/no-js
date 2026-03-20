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
      appcore/                        # provides app-specific contracts for generated code; required today
        context.go                    # provides appcore.Context; required today
        view_models.go                # provides appcore page view types and RootLayoutView; required today
        loaders.go                    # provides page loaders and metadata helpers; usually required
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
- `page.templ` view types must be `appcore.*`.
- Layout and error/not-found contracts currently depend on `appcore.RootLayoutView`.
- Generated code imports `internal/web/appcore` and `internal/web/resolvers` from the consuming module.
- Route-local `components/` directories are rejected by the generator.
- Only `root.templ` may contain document-level tags like `<html>`, `<head>`, and `<body>`.

## What We Support

- `framework/engine`
  Route execution, concurrent metadata and page loading, layout composition, and streaming root-layout rendering.
- `framework/httpserver`
  HTTP server integration, cache policies, gzip, `/healthz`, static asset mounting, and optional public-file middleware.
- `framework/metagen`
  Canonical URLs, alternate languages/types, robots tags, Open Graph, Twitter cards, Pinterest tags, and HTMX head patch generation.
- `framework/i18n`
  Locale config, locale-aware path handling, request locale context, and routing prefix modes: `always`, `as-needed`, `never`.
- `framework/staticassets`
  Minification, hashing, manifest generation, and versioned asset URLs under `/_assets/`.
- `framework/approutegen`
  Route discovery and generated registry/resolver contracts from the file tree.
- `framework/templgen`
  `templ` generation for selected files or paths.
- `framework/cmd/*`
  CLI entrypoints for route generation, `templ` generation, i18n key generation, and static asset building.

## Nuances

- This framework is intentionally not generic yet. The generator still assumes an `appcore` package and specific template signatures.
- The generator is module-aware: framework imports point to `github.com/RevoTale/no-js`, but generated app imports are resolved from the consuming app's `go.mod`.
- i18n locales are currently normalized to two-letter lowercase codes.
- HTMX support is request-driven. Partial requests are detected through `HX-Request`, and metadata patches are emitted through response headers.
- Static assets and public files are separate concerns:
  `/_assets/` is for fingerprinted build output, while public files are served as fixed root-level paths.

## Development

```bash
task fix
task validate
task test
```

## Origin

`no-js` originated as an extraction from [RevoTale/blog](https://github.com/RevoTale/blog).
