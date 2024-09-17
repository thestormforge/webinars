# Setup 

## Metric Server

What is [Metrics Server](https://github.com/kubernetes-sigs/metrics-server)? 

"Metrics Server collects resource metrics from Kubelets and exposes them in Kubernetes apiserver through Metrics API for use by Horizontal Pod Autoscaler and Vertical Pod Autoscaler. Metrics API can also be accessed by kubectl top, making it easier to debug autoscaling pipelines."

### Via Command line

```
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

```

If using `kind` or `minikube`, add `--kubelet-insecure-tls` as arg:

```
kubectl edit deploy metrics-server -n kube-system
```

### Via Helm

```
# add the metrics-server repo
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/

# install the chart
helm upgrade --install metrics-server metrics-server/metrics-server
```

## Using manifest from this repo

There is a manifest file on this repo, ready to be applied on a kind or minikube cluster.

```
kubectl apply -f metrics-server.yaml
```


## Internals of Metrics Server

It is a controller that serves multiple K8s API endpoints `/apis/metrics/v1beta1`. Control plane talks to metric server over port `10250`.
It registers `metrics.k8s.io` API Group and resources `pods` and `nodes` and generate node from scraping `kubelet` cAdvisor metrics.


Logs
```sh
I0827 17:55:17.132120       1 handler.go:275] Adding GroupVersion metrics.k8s.io v1beta1 to ResourceManager
I0827 17:55:17.240256       1 secure_serving.go:213] Serving securely on [::]:10250
```

API Groups and Resources

```sh
kubectl get pods.metrics.k8s.io -A

kubectl get nodes.metrics.k8s.io

```

## Internals of `APIService`

Kubernetes API allows a controller to register as an API Service. This is the case when you install the metrics server. That's how Kubernetes API knows how to redirect the API REST calls to the metrics server.

```sh
% kubectl get APIService | grep metrics
v1beta1.metrics.k8s.io                       kube-system/metrics-server   True        20h
```

On [lab2](../lab2/README.md), we will see how KEDA register as an external metric server.