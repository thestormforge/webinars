# lab1 - Simple HPA


## Setup Workload

Setting up workloads 
```
 kubectl apply -k cpu/
 kubectl get deploy -n hpa-cpu-demo
```

## Create HPA

```
kubectl -n hpa-cpu-demo autoscale deployment c1-no-requests-no-limits --cpu-percent=50 --min=1 --max=5
kubectl -n hpa-cpu-demo autoscale deployment c2-requests-no-limits --cpu-percent=50 --min=1 --max=5
kubectl -n hpa-cpu-demo autoscale deployment c2-small-requests-no-limits --cpu-percent=50 --min=1 --max=5
```

## Observe HPAs

### HPA Not-Functional

Lets check the workload without requests first.

```sh
% k get hpa c1-no-requests-no-limits -n hpa-cpu-demo 
NAME                       REFERENCE                             TARGETS         MINPODS   MAXPODS   REPLICAS   AGE
c1-no-requests-no-limits   Deployment/c1-no-requests-no-limits   <unknown>/50%   1         5         1          114m

```

Why does HPA show unknown? metrics server is not running?

```sh
k describe hpa c1-no-requests-no-limits -n hpa-cpu-demo
```

### HPA Functional

Now let's check the workload with 500m requests and no limits.

```sh
k describe hpa c2-requests-no-limits -n hpa-cpu-demo
```

Let's inflict CPU utilization on the workload, 450m remembering the workload is requesting 500m.

```
kubectl -n hpa-cpu-demo port-forward svc/requests-no-limits 8082:8080 &
curl --data "millicores=450&durationSec=3600" http://localhost:8082/ConsumeCPU
```

Now, let's check the HPA:

```sh
k get hpa c2-requests-no-limits -n hpa-cpu-demo -w
```

What do you see?

Example of the HPA `.spec`, what is the `scaleTargetRef`?

```yaml
spec:
  maxReplicas: 5
  metrics:
  - resource:
      name: cpu
      target:
        averageUtilization: 50
        type: Utilization
    type: Resource
  minReplicas: 1
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: c2-requests-no-limits
```

Please note that HPA takes some time to scale down even if the CPU is no longer running hot, it is because there is defaults for the stabilization windows, which can be fine tuned under [`behavior`](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#default-behavior)


### Optional: Show the inconsistency of HPA over CPU

Look at the c3-small-requests workload, it has 50millicores as request.

```
kubectl -n hpa-cpu-demo port-forward svc/small-requests-no-limits 8083:8080 &
curl --data "millicores=450&durationSec=3600" http://localhost:8083/ConsumeCPU
```

What do you see now? the HPA after awhile will report to 900% utilization. Remember, utilization is a percentage over request.
Overtime, there will be 5 replicas of the pod and showing smaller utilization.
But if you run the following:

```sh
% kubectl top pods -n hpa-cpu-demo
NAMESPACE            NAME                                               CPU(cores)   MEMORY(bytes)   
hpa-cpu-demo         c3-small-requests-no-limits-5b5c5b6765-5snjv             450m         7Mi             
hpa-cpu-demo         c3-small-requests-no-limits-5b5c5b6765-l58ht             0m           0Mi             
hpa-cpu-demo         c3-small-requests-no-limits-5b5c5b6765-n5b98             0m           0Mi             
hpa-cpu-demo         c3-small-requests-no-limits-5b5c5b6765-tlqg8             0m           0Mi             
hpa-cpu-demo         c3-small-requests-no-limits-5b5c5b6765-vbt4k             0m           1Mi             
```

Note that other replicas are doing nothing, only the first one is actually busy. The replicas are not helping at all.

### Ownership between HPA and Workload

Note when HPA object is created, it has a section that configures the target workload and it takes over a subsection of the workload, which is the number of replicas.

We can see this over [`managedFields`](https://kubernetes.io/docs/reference/using-api/server-side-apply/):

```sh
% kubectl get deploy c2-requests-no-limits -n hpa-cpu-demo --show-managed-fields -o yaml | yq '.metadata.managedFields[0]'
```

```yaml
apiVersion: apps/v1
fieldsType: FieldsV1
fieldsV1:
  f:spec:
    f:replicas: {}
manager: kube-controller-manager
operation: Update
subresource: scale

```
