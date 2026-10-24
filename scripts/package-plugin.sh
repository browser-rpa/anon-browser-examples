#!/usr/bin/env bash
set -euo pipefail

root_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
os_name=${GOOS:-$(go env GOOS)}
arch_name=${GOARCH:-$(go env GOARCH)}
target="${os_name}-${arch_name}"
dist_dir="${root_dir}/dist"
stage_dir=$(mktemp -d)
trap 'rm -rf "${stage_dir}"' EXIT

mkdir -p "${stage_dir}/echo-node/bin"
cp "${root_dir}/plugins/echo-node/manifest.json" "${stage_dir}/echo-node/manifest.json"
GOOS="${os_name}" GOARCH="${arch_name}" go build -trimpath -o "${stage_dir}/echo-node/bin/plugin" "${root_dir}/plugins/echo-node"

mkdir -p "${dist_dir}"
archive="${dist_dir}/echo-node-${target}.abp"
rm -f "${archive}"
(cd "${stage_dir}" && zip -q -r "${archive}" echo-node)
printf 'created %s\n' "${archive}"
