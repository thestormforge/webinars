# KEDA

More info on [keda.sh](keda.sh).

Setting up HPA with external metric is possible but with KEDA, a CNCF project is much easier.

KEDA will create and own an HPA object.

We will run a quick demo using an external metric with a rabbit MQ.


## Install KEDA

```sh
helm repo add kedacore https://kedacore.github.io/charts
helm repo update
helm install keda kedacore/keda --namespace keda --create-namespace
```

## Verifying KEDA installation

```sh

# controllers
% k get pods -n keda
NAME                                              READY   STATUS    RESTARTS       AGE
keda-admission-webhooks-85bfd658f5-tvgk5          1/1     Running   1 (4m5s ago)   4m17s
keda-operator-c6f875576-l5567                     1/1     Running   1 (4m8s ago)   4m17s
keda-operator-metrics-apiserver-6d5b8869f-x6qpq   1/1     Running   1 (4m2s ago)   4m17s

# services
% k get service -n keda
NAME                              TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)            AGE
keda-admission-webhooks           ClusterIP   10.96.98.201   <none>        443/TCP            4m29s
keda-operator                     ClusterIP   10.96.30.152   <none>        9666/TCP           4m29s
keda-operator-metrics-apiserver   ClusterIP   10.96.251.19   <none>        443/TCP,8080/TCP   4m29s

# KEDA as an external metrics provider
% kubectl get APIService | grep -i keda | grep external
v1beta1.external.metrics.k8s.io              keda/keda-operator-metrics-apiserver   True        4m47s

```

## Example a ScaledObject

KEDA installs multiple CRDs, but we will look at the heart of it: `ScaledObject`.

It defines the metric to be scaled, authentication and triggers. It creates an HPA object for you.

The `trigger.type` section determines how to scale. The `scaleTargetRef` points to the workload (default is `Deployment` but it can scale statefulsets and CRDs).

```yaml
  spec:
    cooldownPeriod: 30
    maxReplicaCount: 30
    pollingInterval: 5
    scaleTargetRef:
      name: rabbitmq-consumer
    triggers:
    - authenticationRef:
        name: rabbitmq-consumer-trigger
      metadata:
        queueLength: "5"
        queueName: hello
      type: rabbitmq
```

```sh
% k get so  -n default        
NAME                SCALETARGETKIND      SCALETARGETNAME     MIN   MAX   TRIGGERS   AUTHENTICATION              READY   ACTIVE   FALLBACK   PAUSED    AGE
rabbitmq-consumer   apps/v1.Deployment   rabbitmq-consumer         30    rabbitmq   rabbitmq-consumer-trigger   True    False    Unknown    Unknown   4m13s

```

The `scaledObject` creates and manages an HPA to manage the workload, default appending `keda-hpa` on the same of the object.

```sh
% k get so rabbitmq-consumer  -n default -o yaml | yq .status.hpaName
keda-hpa-rabbitmq-consumer
```

```sh
% k get hpa  keda-hpa-rabbitmq-consumer -n default -o yaml | yq .metadata.ownerReferences
- apiVersion: keda.sh/v1alpha1
  blockOwnerDeletion: true
  controller: true
  kind: ScaledObject
  name: rabbitmq-consumer
  uid: ec06eb73-30ee-4758-9e58-67be91fe3e6b
```

```sh
% k get hpa keda-hpa-rabbitmq-consumer -n default -o yaml | yq .spec                    
maxReplicas: 30
metrics:
  - external:
      metric:
        name: s0-rabbitmq-hello
        selector:
          matchLabels:
            scaledobject.keda.sh/name: rabbitmq-consumer
      target:
        averageValue: "5"
        type: AverageValue
    type: External
minReplicas: 1
scaleTargetRef:
  apiVersion: apps/v1
  kind: Deployment
  name: rabbitmq-consumer
```

## Demo with Scaling with RabbitMQ

https://github.com/kedacore/sample-go-rabbitmq


### Not getting Distracted, but we have to install RabbitMQ

Installing RabbitMQ on KinD cluster:

```sh

helm repo add bitnami https://charts.bitnami.com/bitnami

helm install rabbitmq --set auth.username=user --set auth.password=PASSWORD --set volumePermissions.enabled=true bitnami/rabbitmq  --namespace default

```

Checking RabbitMQ installation:

```sh
% k get pods -n default
```

```sh
NAME         READY   STATUS    RESTARTS   AGE
rabbitmq-0   1/1     Running   0          111s
```


### Setting RabbitMQ Consumer with Autoscaling

```sh
kubectl apply -f https://raw.githubusercontent.com/kedacore/sample-go-rabbitmq/main/deploy/deploy-consumer.yaml -n default
```

Checking the deployment, it is scaled to 0 despite `minReplicas: 1` from the HPA object.

```sh
k get deploy rabbitmq-consumer -n default
```

```sh
NAME                READY   UP-TO-DATE   AVAILABLE   AGE
rabbitmq-consumer   0/0     0            0           35s
```

Look at the `ScaledObject` from the previous section. If we set `.spec.minReplicaCount` to 1, the deployment will start. By default, KEDA scales to 0.

The `ScaledObject` creates another KEDA object as well: `TriggerAuthentication` that points to the secret for the authenticatin.


Now, let's set up a Kubernetes job to send 300 messages on the rabbitmq queue and monitor the deployment:

```sh
kubectl apply -f https://raw.githubusercontent.com/kedacore/sample-go-rabbitmq/main/deploy/deploy-publisher-job.yaml
```

```sh
k get deploy rabbitmq-consumer -n default -w
```

```sh
k get deploy rabbitmq-consumer -n default -w
```

```sh
NAME                READY   UP-TO-DATE   AVAILABLE   AGE
rabbitmq-consumer   0/0     0            0           71m
rabbitmq-consumer   0/1     0            0           71m
rabbitmq-consumer   0/1     0            0           71m
rabbitmq-consumer   0/1     0            0           71m
rabbitmq-consumer   0/1     1            0           71m
rabbitmq-consumer   1/1     1            1           71m
rabbitmq-consumer   1/4     1            1           71m
rabbitmq-consumer   1/4     1            1           71m
rabbitmq-consumer   1/4     1            1           71m
rabbitmq-consumer   1/4     4            1           71m
rabbitmq-consumer   2/4     4            2           71m
rabbitmq-consumer   3/4     4            3           71m
rabbitmq-consumer   4/4     4            4           71m
(...)
rabbitmq-consumer   8/8     8            8           71m
rabbitmq-consumer   8/16    8            8           71m
(...)
rabbitmq-consumer   16/16   16           16          72m
rabbitmq-consumer   16/30   16           16          72m
```

Please note how KEDA scaled the replicas, doubling the number.

## Other stuff

KEDA registers as an API Service for the external metrics `v1beta1.external.metrics.k8s.io ` and because of that, only one KEDA can run on the cluster and it conflicts with other external metric providers, such as DataDog.

```sh                                                
% kubectl get APIService | grep metrics
v1beta1.external.metrics.k8s.io              keda/keda-operator-metrics-apiserver   True        2m21s         keda/keda-operator-metrics-apiserver   False (MissingEndpoints)   13s
```
