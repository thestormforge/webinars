# Testing CPU Requests and Limits

You must have completed [lab0](../lab0/README.md) for this demo.

We will be using the Kubernetes [resource-consumer application](https://github.com/kubernetes/kubernetes/tree/master/test/images/resource-consumer), the same app used for K8s end-2-end testing for autoscaling, etc.

## Setting Three Deployments

We will configure three deployments of this application:: one with Best-Effort (`c1`),  another with only requests (`c2`) and another with requests and limits (`c3`) - Burstable:

```
kubectl apply -k cpu/

```

Checking the pods, all should be running on `demo-m02` node:

```
% kubectl get pods -n cpu -o wide
```

Show the CPU requests and limits, confirming settings:

```
kubectl get pods -n cpu -o yaml | yq  '[ .items[] | {"name": .metadata.name, "qosClass": .status.qosClass, "resources": {"requests": {"cpu": .spec.containers[0].resources.requests.cpu}, "limits": {"cpu": .spec.containers[0].resources.limits.cpu} } } ]'

# output
- name: c1-no-requests-no-limits-697885b694-bn4gm
  qosClass: BestEffort
  resources:
    requests:
      cpu: null
    limits:
      cpu: null
- name: c2-requests-no-limits-d8db7db45-9zllz
  qosClass: Burstable
  resources:
    requests:
      cpu: 500m
    limits:
      cpu: null
- name: c3-requests-and-limits-9f8599df5-b2xfj
  qosClass: Burstable
  resources:
    requests:
      cpu: 500m
    limits:
      cpu: "1"

```

On a different terminal, let's port-forward the services for each deployment:

```
kubectl port-forward svc/no-requests-no-limits 8081:8080 -n cpu &
kubectl port-forward svc/requests-no-limits 8082:8080 -n cpu &
kubectl port-forward svc/requests-and-limits 8083:8080 -n cpu &
```

## Sending 450m CPU for each deployment

For one hour, lets have each deployment to consume 450m CPU.

```
curl --data "millicores=450&durationSec=3600" http://localhost:8081/ConsumeCPU
curl --data "millicores=450&durationSec=3600" http://localhost:8082/ConsumeCPU
curl --data "millicores=450&durationSec=3600" http://localhost:8083/ConsumeCPU
```

What do you see from the Grafana view?


## Sending 2000m CPU for requests/no-limits

For 15 seconds lets have each deployment requests/no-limits to consume 2000m CPU.

```
curl --data "millicores=1000&durationSec=15" http://localhost:8082/ConsumeCPU
curl --data "millicores=1000&durationSec=15" http://localhost:8082/ConsumeCPU
```

What do you see from the Grafana view?


## Sending 2000m CPU for requests/limits

For 15 seconds lets have each deployment requests/limits to consume 2000m CPU.

```
curl --data "millicores=1000&durationSec=15" http://localhost:8083/ConsumeCPU
curl --data "millicores=1000&durationSec=15" http://localhost:8083/ConsumeCPU
```

What do you see from the Grafana view?


## Deleting Background processes

Be ready for the next lab and delete the backgroup processes:
```sh
kill %2 %3 %4
```
