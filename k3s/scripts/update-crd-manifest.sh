#!/bin/bash
set -euo pipefail

# Only run on server node
[[ -n "${K3S_URL:-}" ]] && exit 0

dir="${K3S_DATA_DIR:-/var/lib/rancher/k3s}"
manifest="$dir/server/manifests/install-crd.yaml"

[[ ! -f "$manifest" ]] && echo "Manifest not found: $manifest" >&2 && exit 1

echo "Patching manifest: $manifest"
sed -i \
    -e "s|\${MNEMONIC}|${MNEMONIC:-}|g" \
    -e "s|\${NETWORK}|${NETWORK:-}|g" \
    -e "s|\${K3S_TOKEN}|${K3S_TOKEN:-}|g" \
    "$manifest"
