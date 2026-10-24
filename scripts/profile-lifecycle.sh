#!/usr/bin/env bash
set -euo pipefail

api_base=${ANON_BROWSER_API:-http://127.0.0.1:8787}
project_id=${ANON_BROWSER_PROJECT:-local}
name=${1:-"API 示例窗口"}

payload=$(jq -n --arg project_id "$project_id" --arg name "$name" '{project_id:$project_id,name:$name,platform:"mac",identity_template:"mac-default",browser_family:"chrome",browser_version:"stable",languages:["zh-CN","en-US"],timezone:"Asia/Shanghai",proxy_mode:"direct",window_width:1440,window_height:900}')
created=$(curl --fail-with-body -sS -X POST "${api_base}/api/profiles" -H 'Content-Type: application/json' --data "$payload")
printf '%s\n' "$created" | jq .
profile_id=$(printf '%s' "$created" | jq -r '.id // .profile.id')

printf '\nProfiles:\n'
curl --fail-with-body -sS "${api_base}/api/profiles?project_id=${project_id}" | jq .
printf '\nCreated profile: %s\n' "$profile_id"
printf 'Use the returned revision when updating or deleting this profile. The script intentionally does not delete it automatically.\n'
