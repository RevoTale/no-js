# AGENTS.md

## Overview
`no-js` is an opinionated Go framework for server-rendered web applications. Runtime packages live under
`framework/`, build-time packages live under `internal/`, and the supported public CLI entrypoint lives under
`cmd/no-js`.

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

## Working Agreements
- Keep changes scoped to the framework and its tooling.
- Prefer backward-compatible improvements to public packages and CLIs.
- If editing this repository inside a larger checkout, also follow the parent instructions in [../AGENTS.md](../AGENTS.md).

## Taskfile Workflow
- `task fix`: format Go sources.
- `task validate`: run `golangci-lint` and deadcode checks.
- `task test`: run validation, then `go test ./...`.
