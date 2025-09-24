## usage 

> Managed by admin

1. apply the full CRDs/Rols/Manager [manifest](./install.yaml) to the cluster.

```bash
kubectl apply -f https://raw.githubusercontent.com/codescalers/kubecloud/main/crd/dist/install.yaml
```

2. add the manager mnemonic as secret

```bash
kubectl create secret generic threefold-credentials --from-literal=mnemonic="your actual mnemonic phrase here"
```

> Managed by user

3. run your solution server here is [example](./server-example.yaml)

```bash
kubectl apply -f ./server-example.yaml
```

4. update your crd with the desired subdomain & backends urls and apply 

```bash
kubectl apply -f ./crd-example.yaml
```

5. now get your tfgw solution

```bash 
kubectl get tfgw
```