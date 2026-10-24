#!/usr/bin/env bash
set -euo pipefail

api_base=${ANON_BROWSER_API:-http://127.0.0.1:8787}
project_id=${ANON_BROWSER_PROJECT:-local}
run_id=${1:?usage: download-artifact.sh RUN_ID ARTIFACT_ID [OUTPUT_FILE]}
artifact_id=${2:?usage: download-artifact.sh RUN_ID ARTIFACT_ID [OUTPUT_FILE]}
output_file=${3:-"artifact-${artifact_id}"}

curl --fail-with-body -sS -L "${api_base}/api/rpa/runs/${run_id}/artifacts/${artifact_id}?project_id=${project_id}&download=1" -o "$output_file"
printf 'saved %s\n' "$output_file"
