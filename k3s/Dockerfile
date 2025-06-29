FROM ubuntu:24.04
RUN export DEBIAN_FRONTEND=noninteractive && apt-get -qy update && \
    apt-get -qy install wget ssh iproute2 ntp && \
    wget --progress=bar:force:noscroll -O /sbin/k3s https://github.com/k3s-io/k3s/releases/download/v1.33.1+k3s1/k3s && \
    chmod +x /sbin/k3s && \
    wget --progress=bar:force:noscroll -O /sbin/kubectl https://dl.k8s.io/release/v1.33.1/bin/linux/amd64/kubectl && \
    chmod +x /sbin/kubectl && \
    wget --progress=bar:force:noscroll -O /sbin/zinit https://github.com/threefoldtech/zinit/releases/download/v0.2.14/zinit && \
    chmod +x /sbin/zinit && \
    wget https://github.com/threefoldtech/mycelium/releases/download/v0.6.1/mycelium-x86_64-unknown-linux-musl.tar.gz && \
    tar -xvf mycelium-x86_64-unknown-linux-musl.tar.gz && mv mycelium /sbin/ && \
    apt-get -qy remove wget && apt-get -qy autoremove && \
    rm -rf /var/lib/apt/lists/* && rm -rf /build/* && unset DEBIAN_FRONTEND

RUN printf '#!/bin/bash\nexport KUBECONFIG=/etc/rancher/k3s/k3s.yaml' >> /etc/profile.d/setkubeconfig.sh
COPY rootfs /
COPY scripts /scripts
RUN chmod +x /scripts/*
COPY manifests /var/lib/rancher/k3s/server/manifests/

ENTRYPOINT [ "zinit", "init" ]
