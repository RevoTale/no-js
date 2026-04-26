# AGENTS.md

## Overview
`no-js` is an opinionated Go framework for server-rendered web applications. Runtime packages live under
`framework/`, build-time packages live under `internal/`, and the supported public CLI entrypoint lives under
`cmd/no-js`.

Read these first when you need context:

- `README.md`
- `docs/framework/developing-no-js.md`
- `docs/framework/ai-agents.md`

If a task affects the consuming-app contract, also inspect:

- `docs/app/getting-started.md`
- `docs/app/conventions.md`

## Project Structure
```text
<go-repo-root>/
  AGENTS.md
  Taskfile.yml
  cmd/
    no-js/
  framework/
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
    projectlayout/
```

## Strict Rules
- MUST use `golangci-lint` as the Go linter.
- MUST enforce a maximum line length of 120 through `.golangci.yml`.
- MUST run validation and tests through `Taskfile.yml`.
- MUST keep this repository library-focused; do not reintroduce product-specific app code under the root module.
- MUST keep public runtime packages free of imports from `internal/bundler/*`.
- MUST keep generator output module-aware: framework imports come from `github.com/RevoTale/no-js`, while consuming-app
  imports must be derived from the target app module.
- MUST use the canonical public terminology in docs and design discussions: `App Bundle`, `Custom Config`,
  `Site Resolver`, and `Advanced composition`.
- MUST keep public framework config generic; app-specific dependencies and product concepts must stay out.
- MUST not reserve `web/bootstrap` as a contract term or required path.
- MUST keep stable app package directories aligned with their Go package name in source, generated code, fixtures,
  and docs. For example, `web/view` must be `package view`, and `web/components/card` should be `package card`.
  Do not use import aliases to hide package/directory mismatches.
- MUST keep e2e fixture apps under `e2e/testdata` valid on disk. Tests for invalid app shapes must copy a valid
  fixture to a temp directory and break only that temp copy before asserting generation failure.
- MUST treat route-control directories under `web/routes` as routing syntax, not mandatory Go package names.
  `_group__*`, `_slot__*`, `_param__*`, `_catchall__*`, and `_optional_catchall__*` files must still use one clear,
  valid package name per directory.
- MUST document deliberate package/import-path mismatches when introduced. Versioned import paths such as `v2` or
  `v3` may intentionally differ from the declared package name.
- MUST surface invalid app-bundle wiring at startup; framework happy-path APIs must not rely on request-time soft
  failures for missing required context or dependencies.
- MUST keep Client Assets route-static by default. A route bundle is produced from the matched page, layout, or 404
  files plus the colocated assets of every imported `web/components` package reachable from those files. For example,
  if `page.templ` imports `web/components/meter`, the matched route bundle includes `meter.css` and `meter.ts` once,
  even if the component renders conditionally or multiple times. Generated routes inject those route bundles through
  `@metagen.Head(meta)`. This keeps runtime cheap, makes asset inclusion predictable, and avoids head/body ordering
  problems from render-time asset registration. Do not make normal component rendering register assets per call; use
  Advanced composition for exceptional render-precise needs.
- MUST keep `README.md` high-level and task-oriented; field-level contract truth belongs in exported Go types and
  focused reference docs, not long README inventories.
- MUST name app migration docs by source and target version, for example
  `docs/app/migrations/v1.3.0-to-v1.4.0.md`. Use `next` only before the target
  release version exists, and rename it to the concrete higher version after release.
- MUST keep the migration versions linking in the `docs/app/migrations.md` file with higher version on the top. 
- MUST keep the migration guide related only to the `docs/app/migrations` directory and `docs/app/migrations.md`. Only `docs/app/migrations.md` file can point to the cocrete migrations.
- MUST make migration guide titles include both the version range and migration topic.
- MUST make migration guides list concrete refactors: affected symbols/files, old code shape, new code shape, and
  verification commands. Prefer examples over plain text.
- MUST learn the `https://templ.guide/llms.md` before dicussing the new features related to the Go Templ, or if eny knowledge is missing regarding Go Templ.

## Working Agreements
- Keep changes scoped to the framework and its tooling.
- Prefer backward-compatible improvements to public packages and CLIs.
- If editing this repository inside a larger checkout, also follow the parent instructions in [../AGENTS.md](../AGENTS.md).
- Prefer the happy-path mental model of `generated.Bundle(appContext)` consumed by `httpserver.NewApp(...)`.
- Treat advanced composition as app-owned and package-agnostic.
- When improving docs, optimize first-time app development and agent onboarding before adding deeper architecture prose.

## Taskfile Workflow
- `task fix`: format Go sources.
- `task validate`: run `golangci-lint` and deadcode checks.
- `task test`: run validation, then `go test ./...`.
