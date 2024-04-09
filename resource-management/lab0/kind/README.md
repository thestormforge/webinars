# KinD instead of MiniKube

You can try to use KinD (Kubernetes on Docker) instead of minikube.
AFAIK, per https://kind.sigs.k8s.io/docs/user/configuration/#nodes , there is no way to specify resource sizes for the nodes.

Commands:

```
kind create cluster --name=demo --config kind-multi-node.yaml 

kubectl get nodes

kubectl taint nodes demo-control-plane node-role.kubernetes.io/control-plane=:NoSchedule 
```
