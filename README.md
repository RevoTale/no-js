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

## Start Here

For app developers using `no-js`:

- [Getting Started](docs/app/getting-started.md)
  The shortest path from empty app tree to `httpserver.NewApp(...)`.
- [App Conventions](docs/app/conventions.md)
  The strict `web/*` contract, generated outputs, and reserved route files.

For contributors developing `no-js` itself:

- [Developing `no-js`](docs/framework/developing-no-js.md)
  Repository boundaries, main implementation areas, and contributor rules.
- [AI Agents](docs/framework/ai-agents.md)
  A framework-repo reading order and editing guide for agents.

## Using `no-js` In An App

## Core Contract

The route generator is strict. The happy path assumes:

```text
web/routes          route tree
web/generated       generated route modules and App Bundle boundary
web/resolvers       handwritten resolver methods
web/view            runtime contracts used by generated code
web/assets          source bundled static assets
web/assets-build    generated hashed static assets
web/public          fixed-path public files served by convention
```

See [App Conventions](docs/app/conventions.md) for the full layout and route
rules.

## Happy Path

The preferred runtime integration is:

```go
handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
	App:    generated.Bundle(appContext),
	Custom: customConfig,
})
```

Terminology:

```text
Framework Config: generic runtime behavior only
App Bundle: generated route/runtime contract
Custom Config: isolated app-specific hooks
Site Resolver: shared domain and canonical URL policy
Advanced composition: any app-owned package, only when needed
```

Convention defaults:

```text
Assets manifest: web/assets-build/manifest.json
Static prefix: /_assets/
Public files: web/public
Localization: auto-wired when web/i18n exists
```

## Discovery Conventions

Reserved files under `web/routes` return structured discovery data. The
framework owns the transport, endpoint paths, and XML/text rendering:

- `robots.go` returns `discovery.Robots`
- `sitemap.go` returns `[]discovery.SitemapEntry`, plus optional
  `GenerateSitemaps` and `SitemapByID`
- `feed.go` returns `discovery.FeedDocument`

Use the exported structs in `framework/discovery/discovery.go` as the
field-level source of truth.

Specification references:

- Robots Exclusion Protocol: [RFC 9309](https://www.rfc-editor.org/rfc/rfc9309.html)
- RSS 2.0: [RSS Specification](https://www.rssboard.org/rss-specification)
- XML Sitemaps: [Sitemaps XML format](https://www.sitemaps.org/protocol.html)
- Alternate-language sitemap links:
  [Google Search Central hreflang guidance](https://developers.google.com/search/docs/advanced/crawling/localized-versions)
- Image sitemap extensions:
  [Google Search Central image sitemaps](https://developers.google.com/search/docs/crawling-indexing/sitemaps/image-sitemaps)

## Build Config

`no-js` supports an optional root config file named `no-js.bundle.yaml`.

This file is for build-time configuration only. It controls deterministic inputs such as project layout paths,
feature flags used during layout resolution, and static asset build settings.

If `no-js.bundle.yaml` is missing, the CLI uses framework defaults. A missing config file is not an error.

If the file exists, `version: 1` is required and unknown fields are rejected.

YAML values override defaults field by field. Unspecified fields keep their default values.

There are no globally required operational YAML fields. Requirements are resolved from the command and enabled
features. For example, `go.mod`, `web/routes`, and `web/view` must exist, the resolved i18n directory must exist if
i18n routing is enabled, and static asset paths must be valid when running static asset generation.

Invalid YAML is an error.

`no-js.bundle.yaml` must not contain runtime or environment-specific values. Keep process-time configuration such as
listen address, site-resolution policy, API tokens, analytics IDs, cache overrides, advanced asset overrides, and
service wiring in app-owned Go server code.

Example default-equivalent config:

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

static_assets:
  manifest_path: web/assets-build/manifest.json
```

`manifest_path` points to generated static-bundle metadata. It stores the asset hash used by the runtime to construct
the final versioned asset prefix. The happy path uses runtime convention defaults instead of manual asset wiring.

## Nuances

- This framework is intentionally opinionated. The generator assumes the strict `web/*` layout and specific template
  signatures.
- The preferred runtime integration is `generated.Bundle(appContext)` with `httpserver.NewApp(...)`.
- Generated code imports `web/view`, but current view contracts still use the package identifier `runtime`.
- The generator is module-aware: framework imports point to `github.com/RevoTale/no-js`, but generated app imports are resolved from the consuming app's `go.mod`.
- Localization is convention-first: when `web/i18n` exists, the target runtime model auto-wires locale support.
- Advanced composition is supported, but it is not tied to a reserved package or directory name.
- Site and canonical-domain policy should be centralized through a `Site Resolver`.
- i18n locales are currently normalized to two-letter lowercase codes.
- HTMX support is request-driven. Partial requests are detected through `HX-Request`, and metadata patches are emitted through response headers.
- Static assets and public files are separate concerns:
  `/_assets/` is the default runtime prefix for fingerprinted build output, while public files are served as fixed
  request paths.

## Origin

`no-js` originated as an extraction from [RevoTale/blog](https://github.com/RevoTale/blog).
