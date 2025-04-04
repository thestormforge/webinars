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
kubectl -n hpa-cpu-demo autoscale deployment c3-small-requests-no-limits --cpu-percent=50 --min=1 --max=5
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

#### The one with large requests

Now let's check the workload with 500m requests and no limits.

```sh
k describe hpa c2-requests-no-limits -n hpa-cpu-demo
```

Please note that the workload is scale to 1, which is the minimum configured:

```
Conditions:
  Type            Status  Reason               Message
  ----            ------  ------               -------
  AbleToScale     True    ScaleDownStabilized  recent recommendations were higher than current one, applying the highest recent recommendation
  ScalingActive   True    ValidMetricFound     the HPA was able to successfully calculate a replica count from cpu resource utilization (percentage of request)
  ScalingLimited  False   DesiredWithinRange   the desired count is within the acceptable range <---------------------------

```

Now, let's inflict CPU utilization for 5 minutes on the workload: 450m CPU utilization. Always remembering the workload is requesting 500m.

```
kubectl -n hpa-cpu-demo port-forward svc/requests-no-limits 8082:8080 &
curl --data "millicores=450&durationSec=3600" http://localhost:8082/ConsumeCPU
```

Let's check the HPA:

```sh
k get hpa c2-requests-no-limits -n hpa-cpu-demo -w
```

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

What do you see?

First, let's check the pod CPU utilization:

```
% kubectl top pods -n hpa-cpu-demo
NAME                                           CPU(cores)   MEMORY(bytes)   
c1-no-requests-no-limits-7ff7b45794-9bt8r      1m           6Mi             
c2-requests-no-limits-76944c4cc7-lx28p         450m         14Mi            
c3-small-requests-no-limits-669d8c4455-sw8zk   1m           6Mi        
```

Answer: 450m of 500m is 90% utilization of the request of a single pod. Since our target is `averageUtilization: 50`, HPA will scale another replica and do the math again.

So HPA will scale another replica:

```
% k describe hpa c2-requests-no-limits -n hpa-cpu-demo
(...)
  Type    Reason             Age   From                       Message
  ----    ------             ----  ----                       -------
  Normal  SuccessfulRescale  60s   horizontal-pod-autoscaler  New size: 2; reason: cpu resource utilization (percentage of request) above target
```

And HPA does the math again:

```
% k get hpa c2-requests-no-limits -n hpa-cpu-demo
NAME                    REFERENCE                          TARGETS        MINPODS   MAXPODS   REPLICAS   AGE
c2-requests-no-limits   Deployment/c2-requests-no-limits   cpu: 45%/50%   1         5         2          10m
```

Let's see the utilization of the second pod:

```
% kubectl top pods -n hpa-cpu-demo                    
NAME                                           CPU(cores)   MEMORY(bytes)   
c1-no-requests-no-limits-7ff7b45794-9bt8r      0m           6Mi             
c2-requests-no-limits-76944c4cc7-lx28p         450m         14Mi            
c2-requests-no-limits-76944c4cc7-vgn7p         0m           2Mi             
c3-small-requests-no-limits-669d8c4455-sw8zk   0m           6Mi      
```

Because I did not inflict more CPU utilization on this workload, the second pod is idle but it reserved (allocated) another 450m for a single pod on the node.

Moral of the story for the large request: knowing the nature of the CPU utilization of the workload - in this case 450m - it scale only one replica, which is acceptable.


Please note that HPA takes some time to scale down even if the CPU is no longer running hot, it is because there is defaults for the stabilization windows, which can be fine tuned under [`behavior`](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#default-behavior)


#### The one with small requests

This lab is very similar with the one before, but the goal of this lab is showing the impact of an incorrect of vertical pod sizing can spawn multiple copies (horizontal) of the workload without necessarily helping.

Look at the c3-small-requests workload, it has 50millicores as request.
Let's inflict the same CPU utilization of the previous section, which is 450m CPU.

```
kubectl -n hpa-cpu-demo port-forward svc/small-requests-no-limits 8083:8080 &
curl --data "millicores=450&durationSec=3600" http://localhost:8083/ConsumeCPU
```

What do you see now? 

Right off the bat, the HPA after awhile will report to 900% utilization. Remember, utilization is a percentage over request.

```
% kubectl top pods -n hpa-cpu-demo
NAME                                           CPU(cores)   MEMORY(bytes)   
c1-no-requests-no-limits-7ff7b45794-9bt8r      0m           6Mi             
c2-requests-no-limits-76944c4cc7-lx28p         451m         14Mi            
c2-requests-no-limits-76944c4cc7-vgn7p         0m           2Mi             
c3-small-requests-no-limits-669d8c4455-sw8zk   419m         14Mi  
```

HPA:

```
% k get hpa c3-small-requests-no-limits -n hpa-cpu-demo
NAME                          REFERENCE                                TARGETS         MINPODS   MAXPODS   REPLICAS   AGE
c3-small-requests-no-limits   Deployment/c3-small-requests-no-limits   cpu: 838%/50%   1         5         1          20m
```


Overtime, there will be 5 replicas of the pod and showing smaller utilization.


```sh
% kubectl top pods -n hpa-cpu-demo
NAMESPACE            NAME                                               CPU(cores)   MEMORY(bytes)   
hpa-cpu-demo         c3-small-requests-no-limits-5b5c5b6765-5snjv             450m         7Mi             
hpa-cpu-demo         c3-small-requests-no-limits-5b5c5b6765-l58ht             0m           0Mi             
hpa-cpu-demo         c3-small-requests-no-limits-5b5c5b6765-n5b98             0m           0Mi             
hpa-cpu-demo         c3-small-requests-no-limits-5b5c5b6765-tlqg8             0m           0Mi             
hpa-cpu-demo         c3-small-requests-no-limits-5b5c5b6765-vbt4k             0m           1Mi             
```

The HPA will show now 225% target "utilization".

```
% k get hpa c3-small-requests-no-limits -n hpa-cpu-demo
NAME                          REFERENCE                                TARGETS         MINPODS   MAXPODS   REPLICAS   AGE
c3-small-requests-no-limits   Deployment/c3-small-requests-no-limits   cpu: 225%/50%   1         5         5          21m
```

But in reality, the other pods are doing nothing, only the first one is actually busy. The extra four replicas are not helping at all. HPA created 4 extra pods, when possibly only one would be enough. Had I configured this HPA to be max replicas to 10, the problem would be even be compounded: it would start even more pods without necessarily carrying the weight, taking more pods on the cluster, allocate CPU shares on the node, etc.

Moral of the story: the vertical size is critical for HPA. If tuned too small, it cause the yoyo effect of multiple replicas without need (along other manifestations).

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
