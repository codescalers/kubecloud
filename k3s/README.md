# Building
cd to k3s directory
`docker build -t threefoldtech/k3s:latest .`

## Running
for running master node
```bash
docker run -it --name master -e K3S_URL=""  -e K3S_TOKEN="<TOKEN>" --privileged threefoldtech/k3s:latest
```
for running a worker node
```bash
docker run -it --name worker -e K3S_URL="https://<MASTER_IP>:6443" -e K3S_TOKEN="<TOKEN>" --privileged threefoldtech/k3s:latest
```

## Flist
https://hub.grid.tf/samehabouelsaad.3bot/abouelsaad-k3s_1.26.0-latest.flist

## entrypoint 

```
zinit init 
```

## ENV Vars
-  `K3S_URL`: For the master node this should be empty for worker nodes should be the master url for example `https://<MASTER_IP>:6443`
- `K3S_TOKEN`: the token for your cluster should be same for all nodes
- `K3S_DATA_DIR`: Data dir for kubernetes default is `/var/lib/rancher/k3s/`
- `K3S_FLANNEL_IFACE`: Interface used by flannel default is `eth0`
- `K3S_DATASTORE_ENDPOINT`: For k3s external data store like etcd, sqlite, postgres or mysql ...
- `K3S_NODE_NAME`: sets node name