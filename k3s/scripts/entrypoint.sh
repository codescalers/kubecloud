#!/bin/bash

mount --make-shared /run
mount --make-shared /sys

if [ ! -z "${K3S_DATA_DIR}" ]; then
    echo "k3s data-dir set to: $K3S_DATA_DIR"
    cp -r /var/lib/rancher/k3s/* $K3S_DATA_DIR && rm -rf /var/lib/rancher/k3s
    EXTRA_ARGS="--data-dir $K3S_DATA_DIR --kubelet-arg=root-dir=$K3S_DATA_DIR/kubelet"
fi
unset K3S_FLANNEL_IFACE

if [[ "${DUAL_STACK}" = "true" && "${MASTER}" = "true" ]]; then
    EXTRA_ARGS="$EXTRA_ARGS --cluster-cidr=10.42.0.0/16,2001:cafe:42::/56"
    EXTRA_ARGS="$EXTRA_ARGS --service-cidr=10.43.0.0/16,2001:cafe:43::/112"
    EXTRA_ARGS="$EXTRA_ARGS --flannel-backend=none --disable-network-policy"
fi

if [ -z "${K3S_URL}" ]; then
    # Add additional SANs for planetary network IP, public IPv4, and public IPv6  
    # https://github.com/threefoldtech/tf-images/issues/98
    ifaces=( "tun0" "eth0" "eth1" "eth2" )
    
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
        exec k3s server  $EXTRA_ARGS 2>&1
elif [ "${MASTER}" = "true" ]; then
    exec k3s server --server $K3S_URL $EXTRA_ARGS 2>&1
else
    exec k3s agent $EXTRA_ARGS 2>&1
fi
