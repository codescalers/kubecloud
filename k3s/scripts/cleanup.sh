#!/bin/bash
#
# Cleanup script to revert both create_flannel_iface.sh and entrypoint.sh changes
# also removes all k3s data and containerd caches
# NOTE: This script is intended for use in a development environment
set -e

bridge="flannel-br"
eth_iface="eth0"

echo "🧹 Starting comprehensive cleanup of K3s and flannel configuration..."
echo ""

# ============================================================================
# PART 1: Stop K3s Services
# ============================================================================
echo "📛 Stopping K3s services..."

if systemctl is-active --quiet k3s; then
    echo "[*] Stopping k3s service..."
    systemctl stop k3s || true
fi

if systemctl is-active --quiet k3s-agent; then
    echo "[*] Stopping k3s-agent service..."
    systemctl stop k3s-agent || true
fi

# Kill any remaining k3s processes
if pgrep -x "k3s" > /dev/null; then
    echo "[*] Killing remaining k3s processes..."
    pkill -9 k3s || true
fi

# Wait a moment for processes to terminate
sleep 2

# ============================================================================
# PART 2: Unmount containerd/k3s filesystems
# ============================================================================
echo ""
echo "💾 Unmounting containerd and k3s related mounts..."

# Function to unmount all mounts under a directory
unmount_recursive() {
    local base_dir="$1"
    if [ -d "$base_dir" ]; then
        # Get all mount points under this directory (sorted by depth, deepest first)
        local mounts=$(mount | grep "$base_dir" | awk '{print $3}' | sort -r)
        
        if [ ! -z "$mounts" ]; then
            echo "[*] Found mounts under $base_dir, unmounting..."
            echo "$mounts" | while read -r mount_point; do
                echo "    ➤ Unmounting $mount_point..."
                umount -l "$mount_point" 2>/dev/null || umount -f "$mount_point" 2>/dev/null || true
            done
            # Wait a moment for unmounts to complete
            sleep 1
        fi
    fi
}

# Unmount in order of dependency
unmount_recursive "/run/k3s/containerd"
unmount_recursive "/run/k3s"
unmount_recursive "/var/lib/rancher/k3s/agent/containerd"
unmount_recursive "/var/lib/rancher/k3s"
unmount_recursive "/var/lib/rancher"
unmount_recursive "/var/lib/kubelet"
unmount_recursive "/run/containerd"

# Also unmount any remaining k3s/kubelet/containerd mounts
k3s_mounts=$(mount | grep -E "k3s|kubelet|containerd" | awk '{print $3}' | sort -r || true)
if [ ! -z "$k3s_mounts" ]; then
    echo "[*] Unmounting remaining k3s-related mounts..."
    echo "$k3s_mounts" | while read -r mount_point; do
        echo "    ➤ Unmounting $mount_point..."
        umount -l "$mount_point" 2>/dev/null || umount -f "$mount_point" 2>/dev/null || true
    done
    sleep 1
fi

# ============================================================================
# PART 3: Clean K3s Data and Containerd Caches
# ============================================================================
echo ""
echo "🗑️  Cleaning K3s data directories and containerd caches..."

# Function to safely remove directory after ensuring unmounts
safe_remove() {
    local dir="$1"
    if [ -d "$dir" ]; then
        echo "[*] Removing $dir..."
        # Try normal removal first
        if ! rm -rf "$dir" 2>/dev/null; then
            # If failed, try to find and unmount any remaining mounts
            local remaining_mounts=$(mount | grep "$dir" | awk '{print $3}' | sort -r)
            if [ ! -z "$remaining_mounts" ]; then
                echo "    ⚠️  Found busy mounts, force unmounting..."
                echo "$remaining_mounts" | while read -r mp; do
                    umount -l "$mp" 2>/dev/null || true
                done
                sleep 1
            fi
            # Try removal again
            rm -rf "$dir" 2>/dev/null || echo "    ⚠️  Some files in $dir may still be busy, will retry..."
        fi
    fi
}

# Standard k3s data directory
safe_remove "/var/lib/rancher/k3s"

# Custom data directory if K3S_DATA_DIR was set (from entrypoint.sh)
if [ ! -z "${K3S_DATA_DIR}" ] && [ -d "${K3S_DATA_DIR}" ]; then
    echo "[*] Removing custom K3S_DATA_DIR: ${K3S_DATA_DIR}..."
    safe_remove "${K3S_DATA_DIR}"
fi

# Containerd cache directories
containerd_dirs=(
    "/var/lib/rancher"
    "/run/k3s"
    "/var/lib/kubelet"
    "/var/lib/cni"
    "/etc/rancher"
    "/etc/cni"
)

for dir in "${containerd_dirs[@]}"; do
    safe_remove "$dir"
done

# Clean up containerd state
safe_remove "/run/containerd"

# Clean up any remaining container images and overlays
if [ -d "/var/lib/docker" ]; then
    echo "[*] Cleaning docker/containerd overlays in /var/lib/docker..."
    rm -rf /var/lib/docker/overlay2/* 2>/dev/null || true
fi

# ============================================================================
# PART 4: Clean CNI Interfaces
# ============================================================================
echo ""
echo "🔌 Cleaning CNI network interfaces..."

# Remove CNI interfaces (created by flannel/k3s)
cni_interfaces=$(ip link show | grep -E "cni0|flannel\.|veth" | awk -F': ' '{print $2}' | cut -d'@' -f1 || true)

if [ ! -z "$cni_interfaces" ]; then
    for iface in $cni_interfaces; do
        echo "[*] Removing CNI interface: $iface"
        ip link delete "$iface" 2>/dev/null || true
    done
fi

# Remove CNI bridge if exists
if ip link show cni0 &> /dev/null; then
    echo "[*] Removing cni0 bridge..."
    ip link delete cni0 2>/dev/null || true
fi

# ============================================================================
# PART 5: Clean iptables rules created by K3s/flannel
# ============================================================================
echo ""
echo "🔥 Cleaning iptables and ip6tables rules..."

# Flush K3s related chains
for table in filter nat mangle; do
    echo "[*] Flushing K3s chains in iptables $table table..."
    iptables -t "$table" -S | grep -E "K3S|CNI|FLANNEL|KUBE" | cut -d ' ' -f 2 | xargs -r -n 1 iptables -t "$table" -F || true
    iptables -t "$table" -S | grep -E "K3S|CNI|FLANNEL|KUBE" | cut -d ' ' -f 2 | xargs -r -n 1 iptables -t "$table" -X || true
    
    echo "[*] Flushing K3s chains in ip6tables $table table..."
    ip6tables -t "$table" -S | grep -E "K3S|CNI|FLANNEL|KUBE" | cut -d ' ' -f 2 | xargs -r -n 1 ip6tables -t "$table" -F || true
    ip6tables -t "$table" -S | grep -E "K3S|CNI|FLANNEL|KUBE" | cut -d ' ' -f 2 | xargs -r -n 1 ip6tables -t "$table" -X || true
done

# ============================================================================
# PART 6: Clean Flannel Bridge Configuration
# ============================================================================
echo ""
echo "🌉 Cleaning flannel bridge configuration..."

# Check if bridge exists
if ! ip link show "$bridge" &> /dev/null; then
  echo "✅ Bridge $bridge does not exist, nothing to clean up"
  exit 0
fi

echo "[*] Detected bridge $bridge, proceeding with cleanup..."

# Step 1: Get current configuration from bridge
ipv4=$(ip -4 addr show dev "$bridge" | awk '$1 == "inet" {print $2}')
ipv4_gw=$(ip route show | awk '$1 == "default" && $5 == "'"$bridge"'" {print $3; exit}')

# Get all IPv6 addresses from bridge (excluding link-local)
mapfile -t ipv6_addrs < <(ip -6 addr show dev "$bridge" | awk '/inet6/ && !/fe80::/ {print $2}')

# Get IPv6 gateway from bridge
ipv6_gw_line=$(ip -6 route show default | grep "dev $bridge" | head -n1)
ipv6_gw=$(echo "$ipv6_gw_line" | awk '{for(i=1;i<=NF;i++) if($i=="via") print $(i+1)}')

# Get the 400::/7 route details
route_400=$(ip -6 route show 400::/7 | grep "dev $bridge" | head -n1)
route_400_gw=$(echo "$route_400" | awk '{for(i=1;i<=NF;i++) if($i=="via") print $(i+1)}')

# Capture all non-default IPv4 routes on bridge
mapfile -t bridge_ipv4_routes < <(ip route show | awk '$1 != "default" && $5 == "'"$bridge"'"')

# Step 2: Find which interfaces are enslaved to the bridge
mapfile -t enslaved_ifaces < <(ip link show master "$bridge" | grep -E "^[0-9]+" | awk -F': ' '{print $2}' | cut -d'@' -f1)

echo "[*] Found enslaved interfaces: ${enslaved_ifaces[*]}"

# Step 3: Determine which interface should get the IPv6 with 400::/7 route
# Look for eth1-eth9 that was likely used before
IPV6_IFACE=""
for iface in "${enslaved_ifaces[@]}"; do
  if [[ "$iface" =~ ^eth[1-9]$ ]]; then
    IPV6_IFACE="$iface"
    echo "[*] Will restore IPv6 configuration to: $IPV6_IFACE"
    break
  fi
done

# Step 4: Remove interfaces from bridge
echo "[*] Removing interfaces from bridge..."
for iface in "${enslaved_ifaces[@]}"; do
  echo "    ➤ Removing $iface from $bridge"
  ip link set "$iface" nomaster
  ip link set "$iface" up
done

# Step 5: Delete the bridge
echo "🗑️  Deleting bridge: $bridge"
ip link set "$bridge" down
ip link delete "$bridge" type bridge

# Step 6: Restore IPv4 configuration to eth0
if [[ -n "$ipv4" ]]; then
  echo "[+] Restoring IPv4 ($ipv4) to $eth_iface"
  ip addr add "$ipv4" dev "$eth_iface"
fi

if [[ -n "$ipv4_gw" ]]; then
  echo "[+] Restoring default IPv4 route via $ipv4_gw on $eth_iface"
  ip route add default via "$ipv4_gw" dev "$eth_iface"
fi

# Step 7: Restore IPv4 routes
echo "[*] Restoring non-default IPv4 routes to $eth_iface..."
for route in "${bridge_ipv4_routes[@]}"; do
  new_route=$(echo "$route" | sed "s/ dev $bridge/ dev $eth_iface/")
  echo "    ➤ $new_route"
  ip route replace $new_route
done

# Step 8: Restore IPv6 configuration
# Determine which IPv6 belongs where based on 400::/7 route
if [[ -n "$route_400_gw" ]]; then
  route_prefix=$(echo "$route_400_gw" | cut -d':' -f1-4)
  
  for ipv6_addr in "${ipv6_addrs[@]}"; do
    addr_prefix=$(echo "$ipv6_addr" | cut -d':' -f1-4 | cut -d'/' -f1)
    
    if [[ "$addr_prefix" == "$route_prefix" && -n "$IPV6_IFACE" ]]; then
      # This IPv6 belongs to the interface with 400::/7 route
      echo "[+] Restoring IPv6 ($ipv6_addr) to $IPV6_IFACE"
      ip addr add "$ipv6_addr" dev "$IPV6_IFACE"
      
      # Restore 400::/7 route
      echo "[+] Restoring 400::/7 route via $route_400_gw on $IPV6_IFACE"
      ip -6 route add 400::/7 via "$route_400_gw" dev "$IPV6_IFACE" || true
    else
      # This IPv6 belongs to eth0
      echo "[+] Restoring IPv6 ($ipv6_addr) to $eth_iface"
      ip addr add "$ipv6_addr" dev "$eth_iface"
    fi
  done
else
  # No 400::/7 route, restore all IPv6 to eth0
  for ipv6_addr in "${ipv6_addrs[@]}"; do
    echo "[+] Restoring IPv6 ($ipv6_addr) to $eth_iface"
    ip addr add "$ipv6_addr" dev "$eth_iface"
  done
fi

# Step 9: Restore default IPv6 route
if [[ -n "$ipv6_gw" ]]; then
  echo "[+] Restoring default IPv6 route via $ipv6_gw on $eth_iface"
  ip -6 route add default via "$ipv6_gw" dev "$eth_iface" || true
fi

# ============================================================================
# PART 7: Clean up systemd services and logs
# ============================================================================
echo ""
echo "📝 Cleaning systemd services and logs..."

# Disable k3s services if they exist
if systemctl list-unit-files | grep -q "k3s.service"; then
    echo "[*] Disabling k3s.service..."
    systemctl disable k3s.service 2>/dev/null || true
fi

if systemctl list-unit-files | grep -q "k3s-agent.service"; then
    echo "[*] Disabling k3s-agent.service..."
    systemctl disable k3s-agent.service 2>/dev/null || true
fi

# Clean k3s logs
log_dirs=(
    "/var/log/pods"
    "/var/log/containers"
)

for log_dir in "${log_dirs[@]}"; do
    if [ -d "$log_dir" ]; then
        echo "[*] Removing logs in $log_dir..."
        rm -rf "$log_dir"
    fi
done

# ============================================================================
# PART 8: Reset sysctl settings (optional, from create_flannel_iface.sh)
# ============================================================================
echo ""
echo "⚙️  Resetting sysctl forwarding settings (optional)..."

# Note: These are set to reasonable defaults, not necessarily the original values
# If you need to preserve original values, comment these out
echo "[*] Resetting IP forwarding settings to defaults..."
sysctl -w net.ipv4.ip_forward=0 2>/dev/null || true
sysctl -w net.ipv6.conf.all.forwarding=0 2>/dev/null || true
sysctl -w net.ipv4.conf.all.proxy_arp=0 2>/dev/null || true
sysctl -w net.ipv6.conf.all.proxy_ndp=0 2>/dev/null || true

if [[ -n "$bridge" ]] && ip link show "$bridge" &> /dev/null; then
    sysctl -w net.ipv6.conf."$bridge".forwarding=0 2>/dev/null || true
    sysctl -w net.ipv4.conf."$bridge".proxy_arp=0 2>/dev/null || true
    sysctl -w net.ipv6.conf."$bridge".proxy_ndp=0 2>/dev/null || true
fi

# ============================================================================
# Summary
# ============================================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Cleanup complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Summary of cleaned components:"
echo "  ✓ K3s services stopped and disabled"
echo "  ✓ K3s data directories removed"
echo "  ✓ Containerd caches cleared"
echo "  ✓ CNI network interfaces removed"
echo "  ✓ iptables/ip6tables rules flushed"
echo "  ✓ Flannel bridge configuration reverted"
echo "  ✓ Network configuration restored"
echo "  ✓ Logs and temporary mounts cleaned"
echo "  ✓ Sysctl settings reset"
echo ""
echo "📊 Current network state:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
ip addr show "$eth_iface" | grep -E "inet |inet6 " || echo "No IPs on $eth_iface"
if [[ -n "$IPV6_IFACE" ]]; then
  echo "---"
  ip addr show "$IPV6_IFACE" | grep -E "inet |inet6 " || echo "No IPs on $IPV6_IFACE"
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "💡 Note: You may want to reboot the system to ensure all changes take effect."
