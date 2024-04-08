# Lab 0 - Setup Minikube with 2 nodes with 2 CPUs each

## Installing

### MacOS
Leveraging `brew` to install components.

```
brew install minikube
brew install qemu
brew install socket_vmnet
```

#### Notes on MacOS

Attention: if you use any VPN, i.e. Cloudflare, stop the VPN before the exercise. Once you are completed, stop the `socket_vmnet` service and restart your VPN.
To stop `socket_vmnet`:
```
sudo brew services stop socket_vmnet
```

As of this date (Apr/2024):
- `minikube` latest version is `1.32.0`.
If you have minikube, you should be able to upgrade it `brew upgrade minikube` or reinstall it `brew reinstall minikube`.
- `qemu` latest version is `8.2.1`.
- `socket_vmnet` latest version is `1.1.4`.

### Linux 

TBD

## Starting minikube

The following command will start `minikube` with `qemu`, two nodes with two vCPUs each. The container-runtime must be `containterd`!:

```
minikube start --driver=qemu --container-runtime=containerd --memory=2048 --cpus=2 --nodes=2 --network=socket_vmnet --profile demokcd 
```

You should get an output similar to this:

```
😄  [demokcd] minikube v1.32.0 on Darwin 14.3.1 (arm64)
✨  Using the qemu2 driver based on user configuration
👍  Starting control plane node demokcd in cluster demokcd
🔥  Creating qemu2 VM (CPUs=2, Memory=2048MB, Disk=20000MB) ...
📦  Preparing Kubernetes v1.28.3 on containerd 1.7.8 ...
❌  Unable to load cached images: loading cached images: stat /Users/rafa/.minikube/cache/images/arm64/registry.k8s.io/etcd_3.5.9-0: no such file or directory
    ▪ Generating certificates and keys ...
    ▪ Booting up control plane ...
    ▪ Configuring RBAC rules ...
🔗  Configuring CNI (Container Networking Interface) ...
🔎  Verifying Kubernetes components...
    ▪ Using image gcr.io/k8s-minikube/storage-provisioner:v5
🌟  Enabled addons: default-storageclass, storage-provisioner

👍  Starting worker node demokcd-m02 in cluster demokcd
🔥  Creating qemu2 VM (CPUs=2, Memory=2048MB, Disk=20000MB) ...
🌐  Found network options:
    ▪ NO_PROXY=192.168.105.8
📦  Preparing Kubernetes v1.28.3 on containerd 1.7.8 ...
    ▪ env NO_PROXY=192.168.105.8
    > kubeadm.sha256:  64 B / 64 B [-------------------------] 100.00% ? p/s 0s
    > kubectl.sha256:  64 B / 64 B [-------------------------] 100.00% ? p/s 0s
    > kubelet.sha256:  64 B / 64 B [-------------------------] 100.00% ? p/s 0s
    > kubeadm:  44.75 MiB / 44.75 MiB [--------------] 100.00% 3.78 MiB p/s 12s
    > kubectl:  45.50 MiB / 45.50 MiB [--------------] 100.00% 3.73 MiB p/s 12s
    > kubelet:  100.31 MiB / 100.31 MiB [------------] 100.00% 3.68 MiB p/s 27s
🔎  Verifying Kubernetes components...
🏄  Done! kubectl is now configured to use "demokcd" cluster and "default" namespace by default
```

### Checking the node resources

The following command should say the worker should have two vCPUs.
```
kubectl describe nodes demokcd-m02
```

What to expect:

```
Name:               demokcd-m02
Roles:              <none>
(...)
Capacity:
  cpu:                2
  ephemeral-storage:  17784760Ki
  hugepages-1Gi:      0
  hugepages-2Mi:      0
  hugepages-32Mi:     0
  hugepages-64Ki:     0
  memory:             2011844Ki
  pods:               110
(...)
```

### Tainting control-plane

This step will be important for the next lab:
```
kubectl taint nodes demokcd node-role.kubernetes.io/control-plane=:NoSchedule
```