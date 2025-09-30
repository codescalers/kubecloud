#!/bin/bash

if [ ! -z "${K3S_DATA_DIR}" ]; then
    echo "k3s data-dir set to: $K3S_DATA_DIR"
    cp -r /var/lib/rancher/k3s/* $K3S_DATA_DIR && rm -rf /var/lib/rancher/k3s
    EXTRA_ARGS="--data-dir $K3S_DATA_DIR --kubelet-arg=root-dir=$K3S_DATA_DIR/kubelet"
fi

if [ -z "${K3S_FLANNEL_IFACE}" ]; then
    K3S_FLANNEL_IFACE=flannel-br
    if [ -z "${DUAL_STACK}" ]; then
    K3S_FLANNEL_IFACE=eth0
    fi
fi

if [[ "${DUAL_STACK}" = "true" && "${MASTER}" = "true" ]]; then
    EXTRA_ARGS="$EXTRA_ARGS --cluster-cidr=10.42.0.0/16,2001:cafe:42::/56"
    EXTRA_ARGS="$EXTRA_ARGS --service-cidr=10.43.0.0/16,2001:cafe:43::/112"
    EXTRA_ARGS="$EXTRA_ARGS --flannel-ipv6-masq"
fi

if [[ "${DUAL_STACK}" = "true" ]]; then
    # this to force the ip selection from flannel-br to use mycelium ip
    # not any other ipv6 on flannel-br

    if [ -z "$K3S_FLANNEL_IFACE" ]; then
        echo "Usage: $0 <interface>"
        exit 1
    fi

    # Step 1: Find the next-hop for 400::/7
    route_line=$(ip -6 route | grep "^400::/7.*dev $K3S_FLANNEL_IFACE")
    if [ -z "$route_line" ]; then
        echo "No 400::/7 route found via interface $K3S_FLANNEL_IFACE"
        exit 1
    fi

    # Extract next-hop IPv6
    nexthop=$(echo "$route_line" | awk '{for(i=1;i<=NF;i++) if ($i=="via") print $(i+1)}')
    prefix=$(echo "$nexthop" | cut -d':' -f1-4)

    # Step 2: Get the IPv4 address
    ipv4=$(ip -4 addr show dev "$K3S_FLANNEL_IFACE" | awk '/inet / {print $2}' | cut -d'/' -f1)

    # Step 3: Get global IPv6 addresses and match subnet
    ipv6_list=$(ip -6 addr show dev "$K3S_FLANNEL_IFACE" scope global | awk '/inet6/ {print $2}' | cut -d'/' -f1)
    ipv6=""

    for ip in $ipv6_list; do
        ip_prefix=$(echo "$ip" | cut -d':' -f1-4)
        if [ "$ip_prefix" = "$prefix" ]; then
            ipv6=$ip
            break
        fi
    done

    EXTRA_ARGS="$EXTRA_ARGS --node-ip=$ipv4,$ipv6"
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
