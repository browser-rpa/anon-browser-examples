#!/usr/bin/env bash
set -euo pipefail

api_base=${ANON_BROWSER_API:-http://127.0.0.1:8787}
curl --fail-with-body -sS "${api_base}/healthz"
printf '\n'
