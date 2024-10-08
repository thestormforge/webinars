clear
go run demo.go --inspect-nodes --auto --auto-timeout 2s
sleep 10
clear
go run demo.go --setup-cpu --auto --auto-timeout 2s 
sleep 10
clear
go run demo.go --cpu --auto --auto-timeout 2s  --immediate
sleep 15

# show grafana

## loading extra cpu - requests
#clear
#.go run demo.go --cpu-requests --auto --auto-timeout 2s  --immediate
#sleep 15

## loading extra cpu - limits
#clear
#go run demo.go --cpu-limits --auto --auto-timeout 2s  --immediate
#sleep 15
