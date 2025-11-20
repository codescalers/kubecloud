#!/bin/bash
set -euo pipefail

# Always patch at the original location BEFORE it gets copied to K3S_DATA_DIR
dir="/var/lib/rancher/k3s"
manifest="$dir/server/manifests/tfgw-crd.yaml"


# If K3S_URL not found, patch the manifest. it is a server node
[[ ! -f "$manifest" ]] && echo "Manifest not found: $manifest" >&2 && exit 1

echo "Patching manifest: $manifest"
sed -i \
    -e "s|\${MNEMONIC}|${MNEMONIC:-}|g" \
    -e "s|\${NETWORK}|${NETWORK:-}|g" \
    -e "s|\${TOKEN}|${TOKEN:-}|g" \
    "$manifest"
