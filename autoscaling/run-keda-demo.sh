go run demo.go --setup-keda --auto
sleep 2
clear
go run demo.go --setup-rabbitmq --auto
sleep 2
clear
go run demo.go --setup-rabbitmq-workload --auto
sleep 2
clear
go run demo.go --show-rabbitmq-workload --auto
sleep 2
clear
go run demo.go --scale-rabbitmq-workload --auto
echo "Waiting 15 seconds..."; sleep 15
clear
go run demo.go --show-rabbitmq-workload --auto
echo "Waiting 15 seconds..."; sleep 15
clear
go run demo.go --show-rabbitmq-workload --auto
echo "Waiting 60 seconds..."; sleep 60
clear
go run demo.go --show-rabbitmq-workload --auto
