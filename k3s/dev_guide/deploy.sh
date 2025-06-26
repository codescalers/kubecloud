root=$1
index=$2
host=$3
disk=$4
disk2=$5
init=$6

bridge="docker0"
tap="cont-${index}"

if ! ip l show $tap > /dev/null; then
    sudo ip tuntap add dev "$tap" mode tap
    sudo ip l set "$tap" master $bridge
    sudo ip l set "$tap" up
fi

socket="/tmp/root.${index}.socket"
sudo virtiofsd-rs --shared-dir ${root} --socket ${socket}  &

if [ -z "${init}" ]; then
    init='init=/sbin/zinit "init"'
fi

exec sudo cloud-hypervisor \
    --kernel kernel \
    --initramfs initramfs-linux.img \
    --console off \
    --serial tty \
    --cpus boot=1 \
    --memory size=2G,shared=on \
    --disk path=$disk path=$disk2\
    --fs tag=/dev/root,socket=${socket}  \
    --net tap=${tap} \
    --cmdline "console=ttyS0 rootfstype=virtiofs vda=/opt/data root=/dev/root host=${host} net_eth0=172.17.0.${index}/24 net_dns=8.8.8.8 net_r4=default,172.17.0.1 rw ${init}" \
    --rng