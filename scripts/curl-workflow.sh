#!/usr/bin/env bash
set -euo pipefail

api_base=${ANON_BROWSER_API:-http://127.0.0.1:8787}
workflow_file=${1:-"$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/workflows/google-search.json"}
key="examples-$(date +%s)-$RANDOM"

curl --fail-with-body -sS -X POST "${api_base}/api/rpa/workflows" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: ${key}" \
  --data-binary "@${workflow_file}"
printf '\n'
