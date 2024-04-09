#!/bin/bash

# Simple commands:
#
# brew install minikube
# brew install qemu
# brew install socket_vmnet
# sudo brew services start socket_vmnet
# minikube start --driver=qemu --container-runtime=containerd --memory=2048 --cpus=2 --nodes=2 --network=socket_vmnet --profile demo
# kubectl taint nodes demo node-role.kubernetes.io/control-plane=:NoSchedule
# kubectl label nodes demo node-role.kubernetes.io/control-plane=
#
# Scripted:

set -e

packages=(minikube socket_vmnet qemu)
for pkg in "${packages[@]}"; do
  brew list "$pkg" || brew install "$pkg"
done

echo "Sudo will be used to check and/or start socket_vmnet."
if ! sudo brew services list | grep socket_vmnet | grep started >/dev/null; then
  sudo brew services start socket_vmnet
fi

minikube start \
  --driver=qemu \
  --container-runtime=containerd \
  --memory=2048 \
  --cpus=2 \
  --nodes=2 \
  --network=socket_vmnet \
  --profile demo

kubectl taint nodes demo node-role.kubernetes.io/control-plane=:NoSchedule
kubectl label nodes demo node-role.kubernetes.io/control-plane=
