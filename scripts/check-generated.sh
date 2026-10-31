#!/usr/bin/env bash
set -euo pipefail
root_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "${root_dir}"
tmp_dir=$(mktemp -d)
trap 'rm -rf "${tmp_dir}"' EXIT
go run ./plugins/echo-node --manifest > "${tmp_dir}/echo.json"
go run ./plugins/capability-node --manifest > "${tmp_dir}/capability.json"
go run ./plugins/managed-publish --manifest > "${tmp_dir}/managed-publish.json"
diff -u plugins/echo-node/manifest.json "${tmp_dir}/echo.json"
diff -u plugins/capability-node/manifest.json "${tmp_dir}/capability.json"
diff -u plugins/managed-publish/manifest.json "${tmp_dir}/managed-publish.json"
go run ./cmd/generate-workflows --check
echo "generated plugin manifests and workflows are up to date"
