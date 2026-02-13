#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXCEPTIONS_FILE="$ROOT_DIR/.coverage-exceptions"
DEFAULT_THRESHOLD="100.0"

declare -A THRESHOLDS

if [[ -f "$EXCEPTIONS_FILE" ]]; then
  while IFS= read -r line; do
    trimmed="${line%%#*}"
    trimmed="${trimmed## }"
    trimmed="${trimmed%% }"
    [[ -z "$trimmed" ]] && continue

    pkg="${trimmed%% *}"
    threshold="${trimmed##* }"
    THRESHOLDS["$pkg"]="$threshold"
  done < "$EXCEPTIONS_FILE"
fi

cd "$ROOT_DIR"

coverage_tmp="$(mktemp)"
trap 'rm -f "$coverage_tmp"' EXIT

failed=0

while IFS= read -r pkg; do
  if [[ "$pkg" == "github.com/mgordon34/kornet-kover" ]]; then
    continue
  fi

  threshold="${THRESHOLDS[$pkg]:-$DEFAULT_THRESHOLD}"
  output="$(go test "$pkg" -cover 2>/dev/null || true)"
  line="${output##*$'\n'}"
  percent="${line##*coverage: }"
  percent="${percent%%% of statements*}"

  if [[ -z "$percent" || "$percent" == "$line" ]]; then
    percent="0.0"
  fi
  if [[ "$percent" == "[no statements]" ]]; then
    percent="0.0"
  fi

  printf "%s %s %s\n" "$pkg" "$percent" "$threshold" >> "$coverage_tmp"
done < <(go list ./...)

echo "Package coverage report"
echo "======================="

while IFS= read -r row; do
  pkg="${row%% *}"
  rest="${row#* }"
  percent="${rest%% *}"
  threshold="${rest##* }"

  meets="$(awk -v p="$percent" -v t="$threshold" 'BEGIN { if (p+0 >= t+0) print "yes"; else print "no" }')"

  if [[ "$meets" == "yes" ]]; then
    echo "PASS $pkg coverage=$percent threshold=$threshold"
  else
    echo "FAIL $pkg coverage=$percent threshold=$threshold"
    failed=1
  fi
done < "$coverage_tmp"

if [[ "$failed" -ne 0 ]]; then
  exit 1
fi
