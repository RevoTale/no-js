# App Migrations

Use this page when upgrading an app between `no-js` versions.

Migration guides are ordered with the highest target version first. Files under
`docs/app/migrations/` are named by source and target version. A `next` target is
temporary and must be renamed to the concrete release version after release.

## Next

- [v1.3.0 to Next: View Package, 404 Constructors, And Server Error UI](migrations/v1.3.0-to-next.md)
  Rename app `web/view` packages to `package view`, refactor typed
  `NewNotFoundView(T)` constructors to `NewNotFoundView()`, and move custom 500
  UI from generated `error.templ` wiring to `httpserver.CustomConfig.ServerErrorPage`.
