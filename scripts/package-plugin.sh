#!/usr/bin/env bash
set -euo pipefail

root_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "${root_dir}"
plugin_name=${1:-echo-node}
if [[ ! "${plugin_name}" =~ ^[a-z0-9][a-z0-9-]*$ ]]; then
  printf 'invalid plugin name\n' >&2
  exit 1
fi
plugin_dir="${root_dir}/plugins/${plugin_name}"
if [[ ! -d "${plugin_dir}" || ! -f "${plugin_dir}/manifest.json" ]]; then
  printf 'unknown plugin %s\n' "${plugin_name}" >&2
  exit 1
fi
os_name=${GOOS:-$(go env GOOS)}
arch_name=${GOARCH:-$(go env GOARCH)}
target="${os_name}-${arch_name}"
dist_dir="${root_dir}/dist"
stage_dir=$(mktemp -d)
trap 'rm -rf "${stage_dir}"' EXIT

mkdir -p "${stage_dir}/${plugin_name}/bin"
# Generate descriptors on the build host, even when the package targets another OS.
GOOS="$(go env GOHOSTOS)" GOARCH="$(go env GOHOSTARCH)" go run "${plugin_dir}" --manifest > "${stage_dir}/${plugin_name}/manifest.json"
GOOS="${os_name}" GOARCH="${arch_name}" go build -trimpath -o "${stage_dir}/${plugin_name}/bin/plugin" "${plugin_dir}"

mkdir -p "${dist_dir}"
archive="${dist_dir}/${plugin_name}-${target}.abp"
rm -f "${archive}"
(cd "${stage_dir}" && zip -q -r "${archive}" "${plugin_name}")
printf 'created %s\n' "${archive}"
