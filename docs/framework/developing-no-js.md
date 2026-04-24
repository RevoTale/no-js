# Developing `no-js`

This guide is for contributors working on the framework itself.

## Repository Shape

```text
no-js/
  cmd/
    no-js/
  framework/
    discovery/
    engine/
    httpserver/
    i18n/
    metagen/
    router/
    staticassets/
  internal/
    bundler/
      approutegen/
      i18nkeygen/
      staticassets/
      templgen/
    filesystem/
    projectlayout/
```

## Boundary Rules

- Public runtime packages live under `framework/*`.
- Build-time implementation lives under `internal/*`.
- Public runtime packages must not import `internal/bundler/*`.
- The generator must stay module-aware: framework imports come from
  `github.com/RevoTale/no-js`, while consuming-app imports are resolved from
  the target app module.
- The supported public CLI entrypoint is `cmd/no-js`, and app-facing docs should prefer `tool` directives plus `go tool no-js ...`.
- Keep the repository library-focused. Do not reintroduce product-specific app
  code here.

## Public Mental Model

The framework should preserve this app-facing integration shape:

```go
handler, err := httpserver.NewApp(httpserver.Config[*view.Context]{
	App: generated.Bundle(appContext),
})
```

That implies:

- the generator produces the `App Bundle`
- the consuming app owns `appContext`
- `Custom Config` stays an escape hatch, not the primary integration path

## Main Implementation Areas

- `framework/httpserver`
  Runtime HTTP assembly, happy-path defaults, and app integration.
- `framework/discovery`
  Transport and serialization for `robots.txt`, sitemap endpoints, and RSS.
- `internal/projectlayout`
  Build-time layout defaults and `no-js.bundle.yaml` resolution.
- `internal/bundler/approutegen`
  Route discovery and generated contracts.
- `internal/bundler/staticassets`
  Asset processing, hashing, and manifest generation.

## Validation Loop

Use the framework Taskfile:

```bash
task fix
task validate
task test
```

## Documentation Rules

- Keep `README.md` focused on using `no-js` in an app.
- Use `docs/app/overview.md` as the routing page for app-consumer docs.
- Keep app-consumer docs under `docs/app/`.
- Keep app reference pages under `docs/app/reference/`.
- Keep symptom-first fixes under `docs/app/troubleshooting.md`.
- Keep framework-contributor docs under `docs/framework/`.
- Put field-level contract truth on exported Go types and focused reference
  docs, not in long README inventories.
