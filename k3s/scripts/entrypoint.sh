#!/bin/bash

if [ ! -z "${K3S_DATA_DIR}" ]; then
    echo "k3s data-dir set to: $K3S_DATA_DIR"
    cp -r /var/lib/rancher/k3s/* $K3S_DATA_DIR && rm -rf /var/lib/rancher/k3s
    EXTRA_ARGS="--data-dir $K3S_DATA_DIR --kubelet-arg=root-dir=$K3S_DATA_DIR/kubelet"
fi

if [ -z "${K3S_FLANNEL_IFACE}" ]; then
    K3S_FLANNEL_IFACE=eth0
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
    exec k3s server --flannel-iface $K3S_FLANNEL_IFACE $EXTRA_ARGS 2>&1
else
    exec k3s agent --flannel-iface $K3S_FLANNEL_IFACE $EXTRA_ARGS 2>&1
fi
