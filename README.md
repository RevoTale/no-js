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
your-app/                            # provides the app root; required
  go.mod                             # provides the module path for generated imports; required
  web/                               # provides the fixed web namespace used by the generator; required
    routes/                          # provides the file-based route tree; required
      root.templ                     # provides the document shell; required
      404.templ                      # provides the root not-found page; required
      error.templ                    # provides the root error page; required
      page.templ                     # provides the / route page; optional unless you serve /
      layout.templ                   # provides the root-segment layout; optional
      note/                          # provides a static route segment; optional
        [slug]/                      # provides a dynamic route segment; optional, but [param] syntax is required
          page.templ                 # provides the /note/:slug page; required for this route
    generated/                       # provides generated route modules, registry output, and the App Bundle boundary
    resolvers/                       # provides handwritten route resolver methods; required
      root.go                        # provides resolver methods for the root route; optional per route
      note_param_slug.go             # provides resolver methods for /note/[slug]; optional per route
    view/                            # provides app-specific contracts for generated code; required today
      context.go                     # provides runtime.Context; required today
      view_models.go                 # provides runtime page view types and RootLayoutView; required today
      loaders.go                     # provides page loaders and metadata helpers; usually required
    components/                      # provides shared templ components; optional
    i18n/                            # provides typed message keys and locale files; optional, auto-wired when present
    assets/                          # provides source static assets; optional
    assets-build/                    # provides generated hashed static assets; generated
    public/                          # provides fixed-path public files; optional, served by convention from web/public
  internal/                          # provides app-private domain and infrastructure code; optional
```

Generated files are written to:

```text
web/generated/                     # provides generated route modules with safe Go package names and the App Bundle boundary
web/resolvers/generated.go         # provides generated route resolver interfaces and param types
```

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

Minimal examples:

```go
func Robots(
	runtime framework.RuntimeContext[*runtime.Context],
	r *http.Request,
) (discovery.Robots, error) {
	return discovery.Robots{
		Rules: []discovery.RobotsRule{
			{UserAgent: "*", Allow: []string{"/"}},
		},
		Sitemaps: []string{"https://example.com/sitemap.xml"},
	}, nil
}
```

```go
func Sitemap(
	runtime framework.RuntimeContext[*runtime.Context],
	r *http.Request,
) ([]discovery.SitemapEntry, error) {
	return []discovery.SitemapEntry{
		{
			URL: "https://example.com/",
			Alternates: map[string]string{
				"en": "https://example.com/",
			},
		},
	}, nil
}
```

```go
func Feed(
	runtime framework.RuntimeContext[*runtime.Context],
	r *http.Request,
) (discovery.FeedDocument, error) {
	return discovery.FeedDocument{
		Title:       "Example Feed",
		Link:        "https://example.com/",
		Description: "Latest entries from Example",
		SelfURL:     "https://example.com/feed.xml",
		Items: []discovery.FeedItem{
			{
				Title: "Hello World",
				Link:  "https://example.com/hello-world",
				GUID:  "https://example.com/hello-world",
			},
		},
	}, nil
}
```

Specification references:

- Robots Exclusion Protocol: [RFC 9309](https://www.rfc-editor.org/rfc/rfc9309.html)
- RSS 2.0: [RSS Specification](https://www.rssboard.org/rss-specification)
- XML Sitemaps: [Sitemaps XML format](https://www.sitemaps.org/protocol.html)
- Alternate-language sitemap links:
  [Google Search Central hreflang guidance](https://developers.google.com/search/docs/advanced/crawling/localized-versions)
- Image sitemap extensions:
  [Google Search Central image sitemaps](https://developers.google.com/search/docs/crawling-indexing/sitemaps/image-sitemaps)

## Conventions

These are not suggestions. They are the target framework contract.

- Routes live under `web/routes`.
- Dynamic segments must use `[param]` directories.
- Route templates must use the exact file names `root.templ`, `layout.templ`, `page.templ`, `404.templ`, and `error.templ`.
- `root.templ` is required at `web/routes/root.templ`.
- Root `404.templ` and root `error.templ` are required.
- Generated code imports `web/view` and `web/resolvers` from the consuming module.
- The preferred runtime integration is `generated.Bundle(appContext)` passed to `httpserver.NewApp(...)`.
- `App Bundle` is the generated app contract; `Custom Config` is the isolated escape hatch for app-owned hooks.
- The current `web/view` package is still expected to use the Go package identifier `runtime`.
- `page.templ` view types must be `runtime.*`.
- Layout and error/not-found contracts currently depend on `runtime.RootLayoutView`.
- `web/bootstrap` is not a reserved contract term; advanced composition may live in any app-owned package.
- Site and canonical-domain policy should be centralized through a `Site Resolver`.
- The framework owns the not-found error contract.
- Route-local `components/` directories are rejected by the generator.
- Only `root.templ` may contain document-level tags like `<html>`, `<head>`, and `<body>`.

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

## What We Support

- Runtime packages under `framework/*`
- `framework/engine`
  Route execution, concurrent metadata and page loading, layout composition, and streaming root-layout rendering.
- `framework/httpserver`
  HTTP server integration, the `NewApp(...)` happy path, cache policies, gzip, `/healthz`, convention-based asset and
  public-file mounting, locale auto-wiring, and isolated app `Custom Config` hooks.
- `framework/metagen`
  Canonical URLs, alternate languages/types, robots tags, Open Graph, Twitter cards, Pinterest tags, and HTMX head patch generation.
- `framework/i18n`
  Locale config, locale-aware path handling, request locale context, and routing prefix modes: `always`, `as-needed`, `never`.
- `framework/staticassets`
  Runtime manifest loading and runtime asset URL composition.

- Build-time implementation under `internal/*`
- `internal/bundler/approutegen`
  Route discovery and generated registry/resolver contracts from the file tree.
- `internal/bundler/i18nkeygen`
  Go key generation from canonical locale message definitions.
- `internal/bundler/staticassets`
  Minification, hashing, manifest generation, and versioned asset bundle assembly.
- `internal/bundler/templgen`
  `templ` generation for selected files or paths.
- `internal/projectlayout`
  Strict layout defaults, config loading, and module-aware path resolution.

- CLI entrypoints
- `cmd/no-js`
  Recommended root CLI for `gen`, `gen routes`, `gen assets`, and `gen check`.
- `cmd/*`
  Low-level compatibility and development CLIs for route generation, `templ` generation, i18n key generation, and
  static asset building. `cmd/no-js` is the supported public entrypoint.

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
- Build-time implementation may import runtime-owned shared types, but public runtime packages must not import
  `internal/bundler/*`.
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
