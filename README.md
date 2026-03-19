# no-js

`no-js` is an opinionated Go framework for server-rendered web applications.

It includes:
- a route execution engine for SSR pages
- metadata and HTML head composition
- locale resolution and i18n middleware
- static asset fingerprinting and manifest generation
- code generators for route wiring, `templ`, and i18n keys

The reusable code lives under `framework/`.

## Conventions

The route generator is intentionally opinionated. It expects the consuming app to keep route templates and resolver code under:

- `internal/web/app`
- `internal/web/appcore`
- `internal/web/resolvers`

Generated code imports framework packages from `github.com/RevoTale/no-js/framework`.

## Development

```bash
task fix
task validate
task test
```
