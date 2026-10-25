# Continuous integration

Woodpecker runs `.woodpecker/` for pushes, pull requests, manual runs, and tags.
`validate` lints workflow syntax. `test` installs pinned tools, then runs the same
`task test` validation/e2e suite used locally.

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

E2E subprocesses inherit the pipeline's Go build cache. Go's source/compiler/flag
cache keys keep fixture builds correct while avoiding repeated standard-library
compilation. Their generated files, binaries, and HTTP state remain separate.
The shared CLI binary is removed after the E2E suite. No host cache volume is used.

## Local race gate

Before merging, run `task perf:race` in a compatible local Dev Container on the
exact source revision being reviewed and record the result in the PR. This is a
mandatory local gate. The current hosted ARM64 runner has a 47-bit VMA which its
Go ThreadSanitizer runtime rejects before tests execute. The owner approved
keeping the full original CI suite hosted and race validation local; do not
silently skip the local gate or report it as a hosted CI pass.
