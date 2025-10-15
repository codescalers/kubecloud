#!/bin/bash
set -euo pipefail

# Wait until the API server is reachable
until [ "$(kubectl get --raw='/readyz')" = "ok" ]; do
    echo "Waiting for Kubernetes API...";     
    sleep 5; 
done

# Only run on leader
if [[ -z "${K3S_URL:-}" ]]; then
    dir="${K3S_DATA_DIR:-/var/lib/rancher/k3s}"
    manifest="$dir/server/manifests/install-crd.yaml"

    if [[ -f "$manifest" ]]; then
        echo "Patching manifest: $manifest"
        sed -i \
            -e "s|\${MNEMONIC}|${MNEMONIC:-}|g" \
            -e "s|\${NETWORK}|${NETWORK:-}|g" \
            -e "s|\${K3S_TOKEN}|${K3S_TOKEN:-}|g" \
            "$manifest"

        echo "Applying manifest..."
        kubectl apply -f "$manifest"
    else
        echo "Manifest not found: $manifest" >&2
        exit 1
    fi
fi
