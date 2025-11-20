## usage 

> Managed by admin

1. update the manifest with your mnemonic

```bash
sed "s|\${MNEMONIC}|$MNEMONIC|g; s|\${NETWORK}|$NETWORK|g s|\${TOKEN}|$TOKEN|g" ./install.yaml

```

2. apply the full CRDs/Rols/Manager [manifest](./install.yaml) to the cluster.

```bash
kubectl apply -f ./install.yaml
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

6. now you need to apply ingress on cluster redirect to your service

```bash
# update the example with the generated FGDN
kubectl apply -f ./ingress-example.yaml
```