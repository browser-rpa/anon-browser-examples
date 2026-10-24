#!/usr/bin/env bash
set -euo pipefail

api_base=${ANON_BROWSER_API:-http://127.0.0.1:8787}
profile_id=${1:?usage: runtime-lifecycle.sh PROFILE_ID}

printf 'Current runtimes:\n'
curl --fail-with-body -sS "${api_base}/api/runtimes" | jq .
printf '\nStart runtime:\n'
curl --fail-with-body -sS -X POST "${api_base}/api/runtimes/${profile_id}/start" \
  -H 'Content-Type: application/json' -d '{"headless":false}' | jq .
printf '\nStop runtime:\n'
curl --fail-with-body -sS -X POST "${api_base}/api/runtimes/${profile_id}/stop" | jq .
