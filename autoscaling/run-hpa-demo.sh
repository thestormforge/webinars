go run demo.go --setup-metrics --auto --immediate
sleep 2
clear
go run demo.go --setup-cpu --auto # goes on the video demo
sleep 2
clear
go run demo.go --load-cpu --auto --immediate
sleep 2
clear
go run demo.go --cpu-top --auto # goes on the video demo, make sure it is showing 450 milli on each pod
sleep 2
clear
go run demo.go --setup-hpa --auto # goes on the video demo
sleep 2
clear
go run demo.go --show-hpa --auto # goes on the video demo, make sure it is showing 90% and 900%
sleep 2
clear
# wait HPA to scale
go run demo.go --show-hpa --auto --immediate # goes on the video demo, make sure it is scaled
sleep 2
clear
go run demo.go --cpu-top-c2 --auto # goes on the video, explain why HPA is now 45%
sleep 2
clear
go run demo.go --cpu-top-c3 --auto # goes on the video, explain why HPA is now 225%
sleep 2
clear
