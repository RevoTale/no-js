# AGENTS.md

## Overview
`no-js` is an opinionated Go framework for server-rendered web applications. The repository is framework-first: reusable
runtime packages and generator CLIs live under `framework/`.

## Project Structure
```text
<go-repo-root>/
  AGENTS.md
  Taskfile.yml
  framework/
    approutegen/
    cmd/
    engine/
    httpserver/
    i18n/
    metagen/
    router/
    staticassets/
    templgen/
```

## Strict Rules
- MUST use `golangci-lint` as the Go linter.
- MUST enforce a maximum line length of 120 through `.golangci.yml`.
- MUST run validation and tests through `Taskfile.yml`.
- MUST keep this repository library-focused; do not reintroduce product-specific app code under the root module.
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
