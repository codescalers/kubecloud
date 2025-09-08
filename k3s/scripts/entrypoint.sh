#!/bin/bash

# Set dual-stack as default if not specified
if [ -z "${DUAL_STACK}" ]; then
    DUAL_STACK="true"
fi

if [ ! -z "${K3S_DATA_DIR}" ]; then
    echo "k3s data-dir set to: $K3S_DATA_DIR"
    cp -r /var/lib/rancher/k3s/* $K3S_DATA_DIR && rm -rf /var/lib/rancher/k3s
    EXTRA_ARGS="--data-dir $K3S_DATA_DIR --kubelet-arg=root-dir=$K3S_DATA_DIR/kubelet"
fi

# Disable Flannel and use Cilium instead
EXTRA_ARGS="$EXTRA_ARGS --flannel-backend=none --disable-network-policy"

if [[ "${DUAL_STACK}" = "true" && "${MASTER}" = "true" ]]; then
    EXTRA_ARGS="$EXTRA_ARGS --cluster-cidr=10.42.0.0/16,2001:cafe:42::/56"
    EXTRA_ARGS="$EXTRA_ARGS --service-cidr=10.43.0.0/16,2001:cafe:43::/112"
    EXTRA_ARGS="$EXTRA_ARGS --service-ip-family-policy=RequireDualStack"
fi

if [[ "${DUAL_STACK}" = "true" ]]; then
    # Detect IPv4 from eth0
    ipv4=$(ip -4 addr show dev eth0 | awk '/inet / {print $2}' | cut -d'/' -f1)
    
    # Find IPv6 interface with 400::/7 route (mycelium network)
    ipv6=""
    for iface in eth{1..9}; do
        if ip -6 route show 400::/7 | grep -qw "dev $iface"; then
            # Get the IPv6 address from this interface that matches the route prefix
            route_line=$(ip -6 route show 400::/7 | grep "dev $iface")
            nexthop=$(echo "$route_line" | awk '{for(i=1;i<=NF;i++) if ($i=="via") print $(i+1)}')
            prefix=$(echo "$nexthop" | cut -d':' -f1-4)
            
            ipv6_list=$(ip -6 addr show dev "$iface" scope global | awk '/inet6/ {print $2}' | cut -d'/' -f1)
            for ip in $ipv6_list; do
                ip_prefix=$(echo "$ip" | cut -d':' -f1-4)
                if [ "$ip_prefix" = "$prefix" ]; then
                    ipv6=$ip
                    break 2
                fi
            done
        fi
    done
    
    if [[ -n "$ipv4" && -n "$ipv6" ]]; then
        EXTRA_ARGS="$EXTRA_ARGS --node-ip=$ipv4,$ipv6"
        echo "Detected node IPs: IPv4=$ipv4, IPv6=$ipv6"
    elif [[ -n "$ipv4" ]]; then
        EXTRA_ARGS="$EXTRA_ARGS --node-ip=$ipv4"
        echo "Detected node IP: IPv4=$ipv4"
    fi
fi 

if [ -z "${K3S_URL}" ]; then
    # Add additional SANs for planetary network IP, public IPv4, and public IPv6  
    # https://github.com/threefoldtech/tf-images/issues/98
    ifaces=( "tun0" "eth1" "eth2" )

    for iface in "${ifaces[@]}"
    do
        addrs="$(ip addr show $iface | grep -E "inet |inet6 "| grep "global" | cut -d '/' -f1 | cut -d ' ' -f6)"
        for addr in $addrs
        do
            # `ip route get` just used here to validate the ip addr to handle edge caese where parsing could misbehave 
            ip route get $addr && EXTRA_ARGS="$EXTRA_ARGS --tls-san $addr"
        done
    done
    if [ "${HA}" = "true" ]; then
        EXTRA_ARGS="$EXTRA_ARGS --cluster-init"
    fi
    exec k3s server --flannel-iface $K3S_FLANNEL_IFACE $EXTRA_ARGS 2>&1
elif [ "${MASTER}" = "true" ]; then
    exec k3s server --server $K3S_URL --flannel-iface $K3S_FLANNEL_IFACE $EXTRA_ARGS 2>&1
else
    exec k3s agent --flannel-iface $K3S_FLANNEL_IFACE $EXTRA_ARGS 2>&1
fi
