#!/bin/bash
set -e

echo "🔧 Setting up networking for Cilium dual-stack"

# Enable IP forwarding for both IPv4 and IPv6
echo "📡 Enabling IP forwarding..."
sysctl -w net.ipv4.ip_forward=1
sysctl -w net.ipv6.conf.all.forwarding=1

# Ensure all interfaces are up and configured
echo "🔍 Checking network interfaces..."

# Check IPv4 interface (eth0)
if ip -4 addr show dev eth0 | grep -q "inet "; then
    ipv4=$(ip -4 addr show dev eth0 | awk '/inet / {print $2}' | cut -d'/' -f1)
    echo "✅ IPv4 interface eth0: $ipv4"
else
    echo "⚠️  No IPv4 address found on eth0"
fi

# Check for IPv6 interface with 400::/7 route
ipv6_iface=""
for iface in eth{1..9}; do
    if ip -6 route show 400::/7 | grep -qw "dev $iface" 2>/dev/null; then
        ipv6_iface=$iface
        ipv6=$(ip -6 addr show dev "$iface" scope global | awk '/inet6/ {print $2}' | cut -d'/' -f1 | head -n1)
        echo "✅ IPv6 interface $iface: $ipv6"
        break
    fi
done

if [[ -z "$ipv6_iface" ]]; then
    echo "⚠️  No IPv6 interface with 400::/7 route found"
fi

# Configure sysctls for optimal Cilium operation
echo "⚙️  Configuring kernel parameters for Cilium..."
sysctl -w net.core.bpf_jit_enable=1
sysctl -w net.core.bpf_jit_harden=0
sysctl -w kernel.unprivileged_bpf_disabled=1

# Ensure netfilter modules are available
echo "🔌 Loading required kernel modules..."
modprobe ip_tables || echo "ip_tables module already loaded or not available"
modprobe ip6_tables || echo "ip6_tables module already loaded or not available"
modprobe xt_socket || echo "xt_socket module already loaded or not available"

echo "✅ Cilium networking setup completed"
