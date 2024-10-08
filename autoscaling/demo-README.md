# `demo.go` code


## HPA Demo

### Video

https://drive.google.com/file/d/1GEFYtfg6tDnHz3DAVZPe1YdtCQ4ELhlS/view?usp=sharing


### Commands

How to run HPA demo:

```
go run demo.go --setup-metrics --auto --immediate
go run demo.go --setup-cpu --auto # goes on the video demo
go run demo.go --load-cpu --auto --immediate
go run demo.go --cpu-top --auto # goes on the video demo, make sure it is showing 450 milli on each pod
go run demo.go --setup-hpa --auto # goes on the video demo
go run demo.go --show-hpa --auto # goes on the video demo, make sure it is showing 90% and 900%
# wait HPA to scale
go run demo.go --show-hpa --auto --immediate # goes on the video demo, make sure it is scaled
go run demo.go --cpu-top-c2 --auto # goes on the video, explain why HPA is now 45%
go run demo.go --cpu-top-c3 --auto # goes on the video, explain why HPA is now 225%
```


## KEDA Demo

```
go run demo.go --setup-keda --auto
go run demo.go --setup-rabbitmq --auto
go run demo.go --setup-rabbitmq-workload --auto
go run demo.go --show-rabbitmq-workload --auto
go run demo.go --scale-rabbitmq-workload --auto
sleep 30
go run demo.go --show-rabbitmq-workload --auto
sleep 15
go run demo.go --show-rabbitmq-workload --auto
sleep 60
go run demo.go --show-rabbitmq-workload --auto
```



## Cleaning 

To clean cluster:

```
go run demo.go --clean-slate --auto --immediate
```