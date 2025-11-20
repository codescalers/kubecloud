#!/bin/bash

set -euo pipefail

EXTRA_ARGS=""

log_info() {
    echo '[INFO] ' "$@"
}

log_fatal() {
    echo '[ERROR] ' "$@" >&2
    exit 1
}

source_env_file() {
    local env_file="${1:-}"
    
    if [ ! -f "$env_file" ]; then
        log_fatal "Environment file not found: $env_file"
    fi
        
    set -a
    source "$env_file"
    set +a
}

check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_fatal "This script must be run as root"
    fi
}

install_deps() {
    log_info "Updating package lists..."
    if ! apt-get update -qq > /dev/null 2>&1; then
        log_fatal "Failed to update package lists"
    fi

    if ! command -v curl &> /dev/null; then
        log_info "Installing curl..."
        apt-get install -y -qq curl > /dev/null 2>&1 || log_fatal "Failed to install curl"
    fi

    if ! command -v ip &> /dev/null; then
        log_info "Installing iproute2 for ip command..."
        apt-get install -y -qq iproute2 > /dev/null 2>&1 || log_fatal "Failed to install iproute2"
    fi

    if ! command -v k3s &> /dev/null; then
        log_info "Installing k3s..."
        if ! curl -fsSL -o /usr/local/bin/k3s https://github.com/k3s-io/k3s/releases/download/v1.33.1+k3s1/k3s 2>/dev/null; then
            log_fatal "Failed to download k3s"
        fi
        chmod +x /usr/local/bin/k3s
    fi

    if ! command -v kubectl &> /dev/null; then
        log_info "Installing kubectl..."
        if ! curl -fsSL -o /usr/local/bin/kubectl https://dl.k8s.io/release/v1.33.1/bin/linux/amd64/kubectl 2>/dev/null; then
            log_fatal "Failed to download kubectl"
        fi
        chmod +x /usr/local/bin/kubectl
    fi
}

get_iface_ipv6() {
    local iface="$1"
    
    # Step 1: Find the next-hop for 400::/7
    local route_line
    route_line=$(ip -6 route | grep "^400::/7.*dev ${iface}" || true)
    if [ -z "$route_line" ]; then
        log_fatal "No 400::/7 route found via interface ${iface}"
    fi

    # Extract next-hop IPv6
    local nexthop
    nexthop=$(echo "$route_line" | awk '{for(i=1;i<=NF;i++) if ($i=="via") print $(i+1)}')
    local prefix
    prefix=$(echo "$nexthop" | cut -d':' -f1-4)

    # Step 3: Get global IPv6 addresses and match subnet
    local ipv6_list
    ipv6_list=$(ip -6 addr show dev "$iface" scope global | awk '/inet6/ {print $2}' | cut -d'/' -f1)

    local ip ip_prefix
    for ip in $ipv6_list; do
        ip_prefix=$(echo "$ip" | cut -d':' -f1-4)
        if [ "$ip_prefix" = "$prefix" ]; then
            echo "$ip"
            return 0
        fi
    done

    log_fatal "No global IPv6 address found on ${iface} matching prefix ${prefix}"
}

prepare_args() {
    log_info "Preparing k3s arguments..."
    
    if [ -z "${K3S_FLANNEL_IFACE:-}" ]; then
        log_fatal "K3S_FLANNEL_IFACE not set, it should be your mycelium interface"
    else 
        local ipv6
        ipv6=$(get_iface_ipv6 "$K3S_FLANNEL_IFACE")
        EXTRA_ARGS="$EXTRA_ARGS --node-ip=$ipv6"
    fi

    if [ -n "${K3S_DATA_DIR:-}" ]; then
        log_info "k3s data-dir set to: $K3S_DATA_DIR"
        if [ -d "/var/lib/rancher/k3s" ] && [ -n "$(ls -A /var/lib/rancher/k3s 2>/dev/null)" ]; then
            cp -r /var/lib/rancher/k3s/* $K3S_DATA_DIR && rm -rf /var/lib/rancher/k3s
        fi
        EXTRA_ARGS="$EXTRA_ARGS --data-dir $K3S_DATA_DIR --kubelet-arg=root-dir=$K3S_DATA_DIR/kubelet"
    fi

    if [[ "${MASTER:-}" = "true" ]]; then
        EXTRA_ARGS="$EXTRA_ARGS --cluster-cidr=2001:cafe:42::/56"
        EXTRA_ARGS="$EXTRA_ARGS --service-cidr=2001:cafe:43::/112"
        EXTRA_ARGS="$EXTRA_ARGS --flannel-ipv6-masq"
    fi

    if [ -z "${K3S_URL:-}" ]; then
        # Add additional SANs for planetary network IP, public IPv4, and public IPv6  
        # https://github.com/threefoldtech/tf-images/issues/98
        local ifaces=( "tun0" "eth1" "eth2" )

        for iface in "${ifaces[@]}"
        do
            # Check if interface exists before querying
            if ! ip addr show "$iface" &>/dev/null; then
                continue
            fi
            
            local addrs
            addrs=$(ip addr show "$iface" 2>/dev/null | grep -E "inet |inet6 " | grep "global" | cut -d '/' -f1 | awk '{print $2}' || true)
            
            local addr
            for addr in $addrs
            do
                # Validate the IP address by trying to route to it
                if ip route get "$addr" &>/dev/null; then
                    EXTRA_ARGS="$EXTRA_ARGS --tls-san $addr"
                fi
            done
        done
        
        if [ "${HA:-}" = "true" ]; then
            EXTRA_ARGS="$EXTRA_ARGS --cluster-init"
        fi
    else
        if [ -z "${K3S_TOKEN:-}" ]; then
            log_fatal "K3S_TOKEN must be set when K3S_URL is specified (joining a cluster)"
        fi
    fi
}

patch_manifests() {
    log_info "Patching manifests..."

    dir="${K3S_DATA_DIR:-/var/lib/rancher/k3s}"
    manifest="$dir/server/manifests/tfgw-crd.yaml"

    # If K3S_URL found, remove manifest and exit. it is an agent node
    if [[ -n "${K3S_URL:-}" ]]; then
        rm -f "$manifest"
        log_info "Agent node detected, removed manifest: $manifest"
        exit 0
    fi

    # If K3S_URL not found, patch the manifest. it is a server node
    [[ ! -f "$manifest" ]] && echo "Manifest not found: $manifest" >&2 && exit 1

    sed -i \
        -e "s|\${MNEMONIC}|${MNEMONIC:-}|g" \
        -e "s|\${NETWORK}|${NETWORK:-}|g" \
        -e "s|\${TOKEN}|${TOKEN:-}|g" \
        "$manifest"
}

run_node() {
    if [ -z "${K3S_URL:-}" ]; then
        log_info "Starting k3s server (initializing new cluster)..."
        log_info "Command: k3s server --flannel-iface $K3S_FLANNEL_IFACE $EXTRA_ARGS"
        exec k3s server --flannel-iface "$K3S_FLANNEL_IFACE" $EXTRA_ARGS 2>&1
    elif [ "${MASTER:-}" = "true" ]; then
        log_info "Starting k3s server (joining existing cluster as master)..."
        log_info "Command: k3s server --server $K3S_URL --flannel-iface $K3S_FLANNEL_IFACE $EXTRA_ARGS"
        exec k3s server --server "$K3S_URL" --flannel-iface "$K3S_FLANNEL_IFACE" $EXTRA_ARGS 2>&1
    else
        log_info "Starting k3s agent (joining existing cluster as worker)..."
        log_info "Command: k3s agent --server $K3S_URL --flannel-iface $K3S_FLANNEL_IFACE $EXTRA_ARGS"
        exec k3s agent --server "$K3S_URL" --flannel-iface "$K3S_FLANNEL_IFACE" $EXTRA_ARGS 2>&1
    fi
}


main() {
    source_env_file "${1:-}"
    check_root
    install_deps
    prepare_args
    patch_manifests
    run_node
}

main "$@"