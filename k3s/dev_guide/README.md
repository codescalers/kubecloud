# Development Guide for K3S image
## This guide will take you through steps for running HA cluster
### Install the required bins 
use this [readme](https://github.com/threefoldtech/cloud-container/blob/main/readme.md) to get the required bins 
cloud-hypervisor, virtiofsd-rs, and build kernel and initramfs 
make sure kernel and initramfs are in current dir and cloud-hypervisor and virtiofs-rs in you path
### Run all the following commands using root user
```bash
sudo -s
```
### Pull and save the image
```bash
docker pull ahmedhanafy725/k3s
docker save ahmedhanafy725/k3s > k3s.tar
```
### Extract root filesystem from the image
```bash
python3 untar.py k3s.tar root_orig
```
### Add your public key to authorized_keys for rootfs 
```bash
mkdir root_orig/root/.ssh/
cat ~/.ssh/id_rsa.pub >> root_orig/root/.ssh/authorized_keys
```
### RUN ETCD container
```bash
docker pull ahmedhanafy725/etcd
docker run -d ahmedhanafy725/etcd
```
### Run setup.sh
```bash
bash setup.sh <http://172.17.0.3:2379> # etcd_endpoint `the ip of the etcd container`
```
- this will create the rootfs for 2 masters and 1 worker
- creates 3 disks each disk will be attached to a machine which will store k8s data
### Run deploy.sh
```bash
bash deploy.sh master1_rootfs 100 master1 disk1.img disk11.img
bash deploy.sh master2_rootfs 101 master2 disk2.img disk22.img
bash deploy.sh worker_rootfs 102 worker disk3.img disk33.img
```
- this will create 3 vms(2 masters and 1 worker)

## Cleaning vms
```bash
bash clean.sh
```
- this will kill the deployed machines and clean rootfs and disks created for them
