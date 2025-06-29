set -ex
cp -rp root_orig master1_rootfs
cp -rp root_orig master2_rootfs
cp -rp root_orig worker_rootfs
etcd_endpoint=$1

echo """K3S_TOKEN=djlksjdla2lewqldlks
K3S_NODE_NAME=master1
K3S_DATA_DIR=/opt/data
K3S_DATASTORE_ENDPOINT=$etcd_endpoint
""" >> master1_rootfs/etc/environment

echo """K3S_TOKEN=djlksjdla2lewqldlks
K3S_NODE_NAME=master2
K3S_DATA_DIR=/opt/data
K3S_DATASTORE_ENDPOINT=$etcd_endpoint
""" >> master2_rootfs/etc/environment

echo """K3S_TOKEN=djlksjdla2lewqldlks
K3S_NODE_NAME=worker
K3S_URL=https://172.17.0.100:6443
K3S_DATA_DIR=/opt/data
""" >> worker_rootfs/etc/environment

truncate --size 5G disk1.img
truncate --size 5G disk11.img
truncate --size 5G disk2.img
truncate --size 5G disk22.img
truncate --size 5G disk3.img
truncate --size 5G disk33.img
mkfs.btrfs disk1.img 
mkfs.btrfs disk2.img
mkfs.btrfs disk3.img
