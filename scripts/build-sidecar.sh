#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET_TRIPLE="${1:-$(rustc -Vv | awk '/^host:/ {print $2}')}"

case "${TARGET_TRIPLE}" in
  x86_64-apple-darwin)
    GOOS="darwin"
    GOARCH="amd64"
    ;;
  aarch64-apple-darwin)
    GOOS="darwin"
    GOARCH="arm64"
    ;;
  *)
    echo "Unsupported sidecar target: ${TARGET_TRIPLE}" >&2
    exit 1
    ;;
esac

OUTPUT_DIR="${ROOT_DIR}/src-tauri/binaries"
OUTPUT_PATH="${OUTPUT_DIR}/structlens-engine-${TARGET_TRIPLE}"
GO_CACHE_DIR="${ROOT_DIR}/.gocache-sidecar"

mkdir -p "${OUTPUT_DIR}"
mkdir -p "${GO_CACHE_DIR}"
GOCACHE="${GO_CACHE_DIR}" CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" \
  go build -o "${OUTPUT_PATH}" "${ROOT_DIR}/cmd/structlens-engine"
