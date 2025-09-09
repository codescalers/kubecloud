#!/bin/bash

# Setup networking prerequisites for Cilium
# This script handles IP forwarding and other kernel parameters
# that Cilium cannot set due to read-only /proc/sys in containers

echo "Setting up networking prerequisites for Cilium..."

# Enable IP forwarding for IPv4 and IPv6
echo 1 > /proc/sys/net/ipv4/ip_forward
echo 1 > /proc/sys/net/ipv6/conf/all/forwarding

# Enable IPv6 on all interfaces
echo 0 > /proc/sys/net/ipv6/conf/all/disable_ipv6
echo 0 > /proc/sys/net/ipv6/conf/default/disable_ipv6

# Configure bridge netfilter (if available)
if [ -f /proc/sys/net/bridge/bridge-nf-call-iptables ]; then
    echo 1 > /proc/sys/net/bridge/bridge-nf-call-iptables
fi

if [ -f /proc/sys/net/bridge/bridge-nf-call-ip6tables ]; then
    echo 1 > /proc/sys/net/bridge/bridge-nf-call-ip6tables
fi

# Load required kernel modules if available
modprobe -q ip_tables || true
modprobe -q ip6_tables || true
modprobe -q xt_socket || true

echo "Networking prerequisites setup completed"
