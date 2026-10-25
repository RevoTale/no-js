#!/bin/sh
set -eu
case "$(uname -m)" in
  x86_64) arch=amd64; checksum=680859dbb4d881a9c72d4d9a8f510825450849af8567deacd7302c01124416fb ;;
  aarch64|arm64) arch=arm64; checksum=d7189d439d4a6058e39e1b2acd7c8b27acd4856691b408f23b242e4294e1df5b ;;
  *) echo 'Unsupported CI architecture' >&2; exit 1 ;;
esac
tool_dir="${CI_TOOLS_DIR:-/usr/local/bin}"
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
curl -fsSL --retry 3 "https://github.com/go-task/task/releases/download/v3.54.0/task_linux_${arch}.tar.gz" -o "$scratch/task.tar.gz"
printf '%s  %s\n' "$checksum" "$scratch/task.tar.gz" | sha256sum -c -
tar -xzf "$scratch/task.tar.gz" -C "$scratch" task
mkdir -p "$tool_dir"
install -m 0755 "$scratch/task" "$tool_dir/task"
