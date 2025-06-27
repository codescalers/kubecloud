#!/bin/bash
sleep 10
mount --make-shared /mnt/data/
mkdir -p /etc/cni
ln -s /mnt/data/agent/etc/cni/net.d /etc/cni
