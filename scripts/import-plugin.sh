#!/usr/bin/env bash
set -euo pipefail

api_base=${ANON_BROWSER_API:-http://127.0.0.1:8787}
package=${1:?usage: import-plugin.sh path/to/plugin.abp}
curl --fail-with-body -sS -X POST "${api_base}/api/plugins/automation/import" -F "file=@${package}" | jq .
