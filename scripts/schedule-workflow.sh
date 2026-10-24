#!/usr/bin/env bash
set -euo pipefail

api_base=${ANON_BROWSER_API:-http://127.0.0.1:8787}
project_id=${ANON_BROWSER_PROJECT:-local}
workflow_id=${1:?usage: schedule-workflow.sh WORKFLOW_ID WORKFLOW_VERSION_ID}
version_id=${2:?usage: schedule-workflow.sh WORKFLOW_ID WORKFLOW_VERSION_ID}
cron=${CRON:-'0 * * * *'}
timezone=${TIMEZONE:-Asia/Shanghai}
key="examples-schedule-$(date +%s)-$RANDOM"

curl --fail-with-body -sS -X POST "${api_base}/api/rpa/workflows/${workflow_id}/schedules" \
  -H 'Content-Type: application/json' -H "Idempotency-Key: ${key}" \
  -d "$(jq -n --arg project_id "$project_id" --arg version_id "$version_id" --arg cron "$cron" --arg timezone "$timezone" '{project_id:$project_id,workflow_version_id:$version_id,cron:$cron,timezone:$timezone,enabled:false,input:{}}')" | jq .
