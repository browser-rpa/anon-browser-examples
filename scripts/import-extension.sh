#!/usr/bin/env bash
set -euo pipefail

api_base=${ANON_BROWSER_API:-http://127.0.0.1:8787}
profile_id=${1:?usage: import-extension.sh PROFILE_ID EXTENSION_ZIP}
archive=${2:?usage: import-extension.sh PROFILE_ID EXTENSION_ZIP}
archive_name=$(basename "$archive")

printf 'The local extension endpoint copies the submitted package into the selected Profile.\n'
curl --fail-with-body -sS -X POST "${api_base}/api/extensions/import/local" \
  -F "profile_id=${profile_id}" -F "paths=${archive_name}" -F "files=@${archive}" | jq .
