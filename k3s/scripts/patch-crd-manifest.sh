#!/bin/bash
set -euo pipefail

dir="${K3S_DATA_DIR:-/var/lib/rancher/k3s}"
manifest="$dir/server/manifests/tfgw-crd.yaml"

# If K3S_URL found, remove manifest and exit. it is an agent node
if [[ -n "${K3S_URL:-}" ]]; then
    rm -f "$manifest"
    exit 0
fi

# If K3S_URL not found, patch the manifest. it is a server node
[[ ! -f "$manifest" ]] && echo "Manifest not found: $manifest" >&2 && exit 1

sed -i \
    -e "s|\${MNEMONIC}|${MNEMONIC:-}|g" \
    -e "s|\${NETWORK}|${NETWORK:-}|g" \
    -e "s|\${TOKEN}|${TOKEN:-}|g" \
    "$manifest"
