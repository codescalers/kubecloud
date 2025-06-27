#!/bin/bash

ip link add name mycelium-br type bridge
BRIDGE_IP=$(ip -6 addr show dev mycelium | awk '/inet6/ && /scope global/ {print $2}' | cut -d/ -f1 | cut -d: -f1-4 | awk '{print $0 "::1/64"}')
ip addr add $BRIDGE_IP dev mycelium-br 
sysctl -w net.ipv6.conf.all.forwarding=1
echo "net.ipv6.conf.br0.proxy_ndp=1" | tee -a /etc/sysctl.conf
echo "net.ipv6.conf.all.proxy_ndp=1" | tee -a /etc/sysctl.conf
sysctl -p
MYCELIUM_IP=$(ip -6 addr show dev mycelium | awk '/inet6/ && /scope global/ {print $2}' | cut -d/ -f1)
ip -6 neigh add proxy $MYCELIUM_IP dev mycelium-br 
