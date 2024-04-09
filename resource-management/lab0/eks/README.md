# Setup on EKS

The cluster is defined in cluster.yaml and created using the following command:

```bash
eksctl create cluster -f cluster.yaml
```

It contains a single m5.large node which has 2 cores and 8 GiB of memory.
