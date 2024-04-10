# Testing Memory Requests and Limits

You must have completed [lab0](../lab0/README.md) for this demo.

We will be using the Kubernetes [resource-consumer application](https://github.com/kubernetes/kubernetes/tree/master/test/images/resource-consumer), the same app used for K8s end-2-end testing for autoscaling, etc.

## Setting Three Deployments

We will configure three deployments of this application:: one with no request, no limits (`m1`),  another with only requests (`m2`) and another with requests and limits (`m3`) with the same amount (Guaranteed):

```
kubectl apply -k memory/

```

Checking the pods, all should be running on `demo-m02` node:

```
% kubectl get pods -n memory -o wide
```

Show the CPU requests and limits, confirming settings:

```
kubectl get pods -n memory -o yaml | yq  '[ .items[] | {"name": .metadata.name, "qosClass": .status.qosClass, "resources": {"requests": {"memory": .spec.containers[0].resources.requests.memory}, "limits": {"memory": .spec.containers[0].resources.limits.memory} } } ]'

# output
- name: m1-no-requests-no-limits-7dc868f7b6-rhzsq
  qosClass: BestEffort
  resources:
    requests:
      memory: null
    limits:
      memory: null
- name: m2-requests-no-limits-5c574844fc-7g222
  qosClass: Burstable
  resources:
    requests:
      memory: 250Mi
    limits:
      memory: null
- name: m3-requests-and-limits-c758fc7f8-cm5kd
  qosClass: Guaranteed
  resources:
    requests:
      memory: 250Mi
    limits:
      memory: 250Mi

```

Check how much of the node has been allocated, see allocation:

```sh
kubectl describe node demo-m02

```

See the Grafana, each pod should be taken around `85Mi`. 


## Setting the baseline of 150Mi Memory for each deployment

But let's have a baseline of 150Mi, which mean each pod during the startup will use this amount of memory:

```sh
kubectl set env deployment/m1-no-requests-no-limits RUNTIME_CONSUME_MEBIBYTES=150 -n memory
kubectl set env deployment/m2-requests-no-limits RUNTIME_CONSUME_MEBIBYTES=150 -n memory
kubectl set env deployment/m3-requests-and-limits RUNTIME_CONSUME_MEBIBYTES=150 -n memory
```
Confirm on Grafana the new memory allocation.

On a different terminal, let's port-forward the services for each deployment:

```
kubectl port-forward svc/no-requests-no-limits 8081:8080 -n memory &
kubectl port-forward svc/requests-no-limits 8082:8080 -n memory &
kubectl port-forward svc/requests-and-limits 8083:8080 -n memory &
```

## Testing Memory Limits

Let's see if memory limits are being respected asking the pod to consume more than its limits, i.e. 500Mi (remember limit is set to 250Mi).

```
curl --data '{"mebibytes": 500, "seconds": 30, "delay": 1}' http://localhost:8083/ConsumeMem
```

What do you see from the Grafana view?

Other commands to run:

```sh
k get pod -l name=requests-and-limits -n memory

# check number of pod restarts
```

```sh

kubectl describe pod -l name=requests-and-limits -n memory | egrep -A 21 '^Containers:$'

# output
    Last State:     Terminated
      Reason:       OOMKilled  <-----------
      Exit Code:    137
      Started:      Tue, 09 Apr 2024 21:15:29 -0500
      Finished:     Tue, 09 Apr 2024 21:19:43 -0500
    Ready:          True
    Restart Count:  1
```

## Checking Node behavior on Memory Overallocation

Let's see how the node and pods behave when pods without limits allocate AN ADDITIONAL memory: 250Mi bringing to the total of 400Mi per pod (trying to simulate a memory leak).

```
curl --data '{"mebibytes": 250, "seconds": 60, "delay": 1}' http://localhost:8081/ConsumeMem
curl --data '{"mebibytes": 250, "seconds": 60, "delay": 1}' http://localhost:8082/ConsumeMem
```

What do you see from the Grafana view? 

```sh
% kubectl get events | grep OOM
5m11s       Warning   SystemOOM                 node/demo-m02   System OOM encountered, victim process: python, pid: 1637
5m5s        Warning   SystemOOM                 node/demo-m02   System OOM encountered, victim process: python, pid: 1795
```

```sh
kubectl describe pod -l name=no-requests-no-limits -n memory |  egrep -A 21 '^Containers:$'
kubectl describe pod -l name=requests-no-limits -n memory |  egrep -A 21 '^Containers:$' 

# output
    Last State:     Terminated
      Reason:       Error   <------------------------ Not OOM
      Exit Code:    137
      Started:      Tue, 09 Apr 2024 21:15:22 -0500
      Finished:     Tue, 09 Apr 2024 21:28:16 -0500

```

## Do requests reserve space

At this time, all three deployments should be in baseline state, which is 150Mi each. Increasing `requests-no-limits` to request 520Mi and use 500Mi.

```sh
kubectl set resources deployment m2-requests-no-limits --requests memory=520Mi  -n memory
kubectl set env deployment/m2-requests-no-limits RUNTIME_CONSUME_MEBIBYTES=500 -n memory
```

What do you see from the Grafana view? 

What if now if you increase the `m2-requests-no-limits` with more memory, what happens with `m1-no-requests-no-limits`?

```sh
curl --data '{"mebibytes": 100, "seconds": 1800, "delay": 1}' http://localhost:8082/ConsumeMem


% kubectl get pods -n memory  | grep m1
m1-no-requests-no-limits-644bff5c86-mk49c   0/1     CrashLoopBackOff   10 (45s ago)   67m
```

## Deleting Background processes

Be ready for the next lab and delete the backgroup processes:
```sh
kill %1 %2 %3 %4
```
