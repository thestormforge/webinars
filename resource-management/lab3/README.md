# Testing Resource Management Policies

You must have completed [lab0](../lab0/README.md) for this demo.

## Setting Up Environment

```sh
kubectl delete --ignore-not-found -k policies/
kubectl create -k policies/namespace.yaml
```

## Showing how LimitRanges Work

Show the LimitRange definition:

```sh
cat policies/default-resources.yaml
```

Show the resource requests and limits defined in a basic deployment

```sh
cat policies/no-requests.yaml
```

Create the basic deployment in the policies namespace:

```sh
kubectl apply -f policies/no-requests.yaml
```

Now apply the LimitRange policy:

```sh
kubectl apply -f policies/default-resources.yaml
```

What happen with the pods from the previous step?

```sh
kubectl get pod -l name=no-requests -n policies -o yaml  |  \
yq '.items[] | .spec.containers[] | {"resources": .resources}'
```

Delete the pod, forcing it to be recreated and check the new settings:

```sh
kubectl delete pod -l name=no-requests -n policies
sleep 1
kubectl get pod -l name=no-requests -n policies -o yaml  |  \
yq '.items[] | .spec.containers[] | {"resources": .resources}'
```

## Showing how ResourceQuotas work

Show the ResourceQuota definition:

```sh
cat policies/default-resources-quota.yaml
```

Show the resource requests and limits defined in a basic deployment:

```sh
cat policies/test-workload.yaml
```

Create the basic deployment in the policies namespace:

```sh
kubectl apply -f policies/test-workload.yaml
```

Show the current requests of all containers in the namespace:

```sh
kubectl get pods -n policies -o yaml | \
yq '[ .items[] | {"name": .metadata.name, "requests": .spec.containers[].resources.requests} ]'
```

Apply the ResourceQuota to the policies namespace:

```sh
kubectl apply -f policies/default-resources-quota.yaml -n policies
```

Show the workload resources for the policies namespace:

```sh
kubectl get deployments,replicasets -n policies; echo;
kubectl get pods -o custom-columns-file=policies/pod-custom-columns.txt -n policies
```

Try scaling the workload up from 2 replicas to 4:

```sh
kubectl scale deployment test-workload --replicas=4 -n policies
```

How did that go?

```sh
kubectl get deployments,replicasets -n policies; echo;
kubectl get pods -o custom-columns-file=policies/pod-custom-columns.txt -n policies
```

Show the quota-related events on the current workload ReplicaSet:

```sh
kubectl describe -n policies $(kubectl get rs -l name=test-workload -n policies -o name | head -1) | \
sed -n '/Events:/, /Name:/ p'
```

What about actual usage? Let's check on Grafana.

