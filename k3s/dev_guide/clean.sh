set -x
pkill -9 cloud

rm -rf master1_rootfs
rm -rf master2_rootfs
rm -rf worker_rootfs
rm -rf disk1.img
rm -rf disk2.img
rm -rf disk3.img