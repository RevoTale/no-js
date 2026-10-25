#!/bin/sh
set -eu
case "$(uname -m)" in
  x86_64) arch=amd64; checksum=2277d43b98ec0054280f2ac26b53268bae97682444678a59a657dd565da021d6 ;;
  aarch64|arm64) arch=arm64; checksum=a2a4e0065aa41be71f7c5ac90f271b61751331e5d04314e62afe4027855f0893 ;;
  *) echo 'Unsupported CI architecture' >&2; exit 1 ;;
esac
tool_dir="${CI_TOOLS_DIR:-/usr/local/bin}"
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
archive="golangci-lint-2.13.2-linux-${arch}"
curl -fsSL --retry 3 "https://github.com/golangci/golangci-lint/releases/download/v2.13.2/${archive}.tar.gz" -o "$scratch/lint.tar.gz"
printf '%s  %s\n' "$checksum" "$scratch/lint.tar.gz" | sha256sum -c -
tar -xzf "$scratch/lint.tar.gz" -C "$scratch" "${archive}/golangci-lint"
mkdir -p "$tool_dir"
install -m 0755 "$scratch/${archive}/golangci-lint" "$tool_dir/golangci-lint"
