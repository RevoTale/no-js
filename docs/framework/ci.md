# Continuous integration

Woodpecker runs `.woodpecker/` for pushes, pull requests, manual runs, and tags.
`validate` lints workflow syntax. `test` installs pinned tools, then runs the same
`task test` validation/e2e suite used locally plus `task perf:race`.

Release automation still uses the existing release-please JSON manifest and
config. After a successful default-branch push pipeline, `release` publishes any
merged release PR, then creates or updates the next release PR. GitHub Actions
is no longer required.

Enable this repository in Woodpecker and select `.woodpecker/` as its config path.
Configure `forge_token` with GitHub repository contents and pull-request write
permissions, scoped to default-branch push release jobs. Do not expose this token
to pull-request test steps. No publication credentials are required for tests.

Install Woodpecker CLI v3.19.0 in the development container and run `task ci:lint`
to check workflows locally. CI image versions and Go tool versions are pinned.
This configuration does not change branch protection: require the Woodpecker
checks before merging once the server has reported their exact check names.

Task and golangci-lint use official release binaries with fixed SHA-256 checksums
for amd64 and arm64, avoiding dependency/toolchain compilation on each CI run.
Only the pinned deadcode checker is built through Go.
