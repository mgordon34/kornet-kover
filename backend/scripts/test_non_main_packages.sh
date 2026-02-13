#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

module_path="$(go list -m)"

packages=()
while IFS= read -r pkg; do
  if [[ "$pkg" == "$module_path" ]]; then
    continue
  fi
  packages+=("$pkg")
done < <(go list ./...)

if [[ "${#packages[@]}" -eq 0 ]]; then
  echo "No non-main packages found to test."
  exit 0
fi

go test "$@" "${packages[@]}"
