#!/usr/bin/env bash
set -euo pipefail

api_base=${ANON_BROWSER_API:-http://127.0.0.1:8787}
project_id=${ANON_BROWSER_PROJECT:-local}
workflow_file=${1:-"$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/workflows/google-search.json"}
key="examples-lifecycle-$(date +%s)-$RANDOM"

created=$(curl --fail-with-body -sS -X POST "${api_base}/api/rpa/workflows" \
  -H 'Content-Type: application/json' -H "Idempotency-Key: ${key}" --data-binary "@${workflow_file}")
printf '%s\n' "$created" | jq .
workflow_id=$(printf '%s' "$created" | jq -r '.workflow.id // .id')
workflow_revision=$(printf '%s' "$created" | jq -r '.workflow.revision // 1')
draft_revision=$(printf '%s' "$created" | jq -r '.draft.revision // 1')

printf '\nValidation:\n'
curl --fail-with-body -sS -X POST "${api_base}/api/rpa/workflows/${workflow_id}/validate" \
  -H 'Content-Type: application/json' --data "$(jq -n --arg project_id "$project_id" '{project_id:$project_id}')" | jq .

settings=$(curl --fail-with-body -sS "${api_base}/api/rpa/workflows/${workflow_id}/settings?project_id=${project_id}")
settings_revision=$(printf '%s' "$settings" | jq -r '.revision // 1')
catalog_revision=$(curl --fail-with-body -sS "${api_base}/api/rpa/node-catalog" | jq -r '.revision // .catalog_revision // 1')

published=$(jq -n --arg project_id "$project_id" --argjson workflow_revision "$workflow_revision" --argjson draft_revision "$draft_revision" --argjson settings_revision "$settings_revision" --argjson catalog_revision "$catalog_revision" '{project_id:$project_id,expected_workflow_revision:$workflow_revision,expected_version_revision:$draft_revision,expected_settings_revision:$settings_revision,catalog_revision:$catalog_revision}')
printf '\nPublish:\n'
published_result=$(curl --fail-with-body -sS -X POST "${api_base}/api/rpa/workflows/${workflow_id}/publish" \
  -H 'Content-Type: application/json' --data "$published")
printf '%s\n' "$published_result" | jq .
version_id=$(printf '%s' "$published_result" | jq -r '.published.id // .published_version_id // empty')

run_payload=$(jq -n --arg project_id "$project_id" --arg version_id "$version_id" '{project_id:$project_id,workflow_version_id:$version_id,input:{},debug:true}')
printf '\nCreate debug Run:\n'
run_result=$(curl --fail-with-body -sS -X POST "${api_base}/api/rpa/workflows/${workflow_id}/runs" \
  -H 'Content-Type: application/json' -H "Idempotency-Key: ${key}-run" --data "$run_payload")
printf '%s\n' "$run_result" | jq .
run_id=$(printf '%s' "$run_result" | jq -r '.run.id // .id')

printf '\nRun events:\n'
curl --fail-with-body -sS "${api_base}/api/rpa/runs/${run_id}/events?project_id=${project_id}&limit=100" | jq .
printf '\nRun: %s\n' "$run_id"
