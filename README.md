# no-js

`no-js` is an opinionated Go framework for server-rendered web applications.
It is built around a strict `web/*` app tree, a generated `App Bundle`, and a
convention-first runtime.

You write templ routes, resolver methods, view models, and app services. `no-js`
generates the routing/runtime glue and lets `httpserver.NewApp(...)` assemble
the default server.

## What You Get

- file-system routing with generated resolver contracts and an `App Bundle`
- `httpserver.NewApp(...)` as the default runtime entrypoint
- metadata and `<head>` composition
- focused guides for optional app capabilities
- build-time configuration through `no-js.bundle.yaml`

## Shortest Path

Start with the standard app tree:

```text
your-app/
  go.mod
  cmd/
    server/
      main.go
  web/
    routes/
      root.templ
      page.templ
      404.templ
    resolvers/
    view/
```

Run generation, implement the generated resolver methods, then pass the
generated `App Bundle` to the runtime:

```bash
go tool no-js gen -root .
```

```go
appContext := view.NewContext(...)

handler, err := httpserver.NewApp(httpserver.Config[*view.Context]{
	App: generated.Bundle(appContext),
})
```

That is the happy path. Reach for `Custom Config` only when the default runtime
wiring is not enough.

For the exact files, use [Getting Started](docs/app/getting-started.md). For
the strict app contract, use [App Conventions](docs/app/conventions.md).

## AI Agent Usage

If you are an AI agent building an application with `no-js`, start with
[AI Agents For App Development](docs/app/ai-agents.md).

If you are modifying the `no-js` framework repository itself, use
[Framework AI Agents](docs/framework/ai-agents.md).

## Client Assets

CSS, JavaScript, and TypeScript can live beside the route or component that uses
them:

```text
web/
  routes/
    page.templ
    page.css
    page.ts
  components/
    meter/
      meter.templ
      meter.css
      meter.ts
```

Generation creates Go helpers beside those files. CSS class names are hashed in
the rendered HTML and bundled CSS. Matched routes automatically receive the CSS
and module scripts for their page, layouts, 404 page, and imported components.

`go tool no-js gen routes -root .` writes the generated Go helpers.
`go tool no-js gen assets -root .` writes the bundled, fingerprinted browser
files. Most apps run both with:

```bash
go tool no-js gen -root .
```

See [Static Assets And Client Assets](docs/app/features/static-assets.md) for
usage and [Asset Pipeline Reference](docs/app/reference/asset-pipeline.md) for
the compile/bundle split.

## Docs

For app developers:

- [App Docs Overview](docs/app/overview.md)
- [AI Agents For App Development](docs/app/ai-agents.md)
- [Getting Started](docs/app/getting-started.md)
- [App Conventions](docs/app/conventions.md)
- [Feature Guides](docs/app/features/overview.md)
- [CLI Reference](docs/app/reference/cli.md)
- [Bundle Config Reference](docs/app/reference/bundle-config.md)
- [HTTP Server Reference](docs/app/reference/httpserver.md)
- [Troubleshooting](docs/app/troubleshooting.md)

For contributors:

- [Developing `no-js`](docs/framework/developing-no-js.md)
- [AI Agents](docs/framework/ai-agents.md)
