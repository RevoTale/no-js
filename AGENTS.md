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
- MUST preserve the strict `web/components` model in the framework:
  * `web/components` contains no files directly.
  * Each component package lives under `web/components/**/<name>/`; nested component packages are allowed. The
    directory basename, Go package name, and component anchor stem must align.
  * For example, `web/components/card/header/header.templ` is valid; `web/components/card/header.templ` is not.
  * Each component package must contain at least one anchor file: `<name>.templ` or `<name>.go`.
  * Handwritten `.templ` files are anchor-only. Split large variants or subparts into another component package
    instead of adding `variants.templ`, `header.templ`, or similar files.
  * Optional component assets must be same-stem: `<name>.css` and at most one
    `<name>.{js,ts,tsx,mjs,mts}`.
  * File extensions are case-sensitive and must be lower-case. For example, `card.css` and `card.tsx` are valid;
    `card.CSS` and `card.TSX` are rejected.
  * Public handwritten Go API belongs only in `<name>.go`; other handwritten non-test `.go` files in the same
    component directory are support files and must not declare exported funcs, methods, types, vars, or consts.
- MUST keep e2e fixture apps under `e2e/testdata` valid on disk. Tests for invalid app shapes must copy a valid
  fixture to a temp directory and break only that temp copy before asserting generation failure.
- MUST treat route-control directories under `web/routes` as routing syntax, not mandatory Go package names.
  `_group__*`, `_slot__*`, `_param__*`, `_catchall__*`, and `_optional_catchall__*` files must still use one clear,
  valid package name per directory.
- MUST document deliberate package/import-path mismatches when introduced. Versioned import paths such as `v2` or
  `v3` may intentionally differ from the declared package name.
- MUST surface invalid app-bundle wiring at startup; framework happy-path APIs must not rely on request-time soft
  failures for missing required context or dependencies.
- MUST keep Client Assets route-static by default. Asset discovery is based on the matched page, layout, slot, or 404
  templates plus the colocated assets of every imported `web/components` package reachable from those files.
- MUST use the layout-subtree CSS mental model for generated Client Asset stylesheets. CSS is folded up to the nearest
  non-root `layout.templ` that contains the subtree, even when that layout has no colocated `layout.css`.
  `web/routes/root.css` stays app-shell-only and must not become "all app CSS" by default. For example,
  `web/routes/dashboard/layout.templ`
  can produce one `routes/dashboard/layout.css` containing descendant page CSS, slot CSS, and imported dashboard
  component CSS, reused by every dashboard page. Page-level CSS is only standalone when there is no non-root layout
  owner that should own that subtree. Final generated CSS files still pass through the static asset builder's esbuild
  CSS transform before reaching the browser; do not treat the pre-static-build staging files as final output.
- MUST keep browser target configuration unified. `assets.browser_targets` defaults to `es2020` and must apply to
  every esbuild-backed path: Client Asset scripts, generated global templ CSS after staging, generated
  route/component CSS after staging, and explicit `web/assets` CSS/JS. Explicit `web/assets` CSS may bundle relative
  CSS `@import` files. Route/component CSS imports must not become dependency discovery; ownership stays based on
  route templates and reachable Go component imports.
- MUST keep JavaScript as shared owner entries plus esbuild chunks. For example, `web/routes/dashboard/page.tsx`
  emits `routes/dashboard/page.js`, component scripts emit under `components/<name>/`, and shared JS imports are
  emitted under `chunks/`. Generated routes inject the ordered owner files they need through `@metagen.Head(meta)`.
  Do not make normal component rendering mutate page assets at render time. Client Assets must be discovered during
  generation from route/layout/slot templates and reachable component imports. Use Advanced composition only when an
  app needs manual, render-precise asset control outside the generated route-static model.
- MUST keep `README.md` high-level and task-oriented; field-level contract truth belongs in exported Go types and
  focused reference docs, not long README inventories.
- MUST layer consuming-app docs for low cognitive load: put the short mental model in getting-started docs, exact
  path examples in feature guides, and full allow-lists or edge-case contracts in conventions/reference docs.
- MUST name app migration docs by source and target version, for example
  `docs/app/migrations/v1.3.0-to-v1.4.0.md`. Use `next` only before the target
  release version exists, and rename it to the concrete higher version after release.
- MUST keep the migration versions linking in the `docs/app/migrations.md` file with the higher version at the top.
- MUST keep the migration guide related only to the `docs/app/migrations` directory and `docs/app/migrations.md`.
  Only `docs/app/migrations.md` file can point to the concrete migrations.
- MUST make migration guide titles include both the version range and migration topic.
- MUST make migration guides list concrete refactors: affected symbols/files, old code shape, new code shape, and
  verification commands. Prefer examples over plain text.
- MUST learn the `https://templ.guide/llms.md` before discussing new features related to Go Templ, or if any
  knowledge is missing regarding Go Templ.

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
