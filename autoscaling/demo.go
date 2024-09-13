package main

import (
	demo "github.com/saschagrunert/demo"
)

func main() {
	d := demo.New()

	d.Name = "Kubernetes autoscaling demo"
	d.Usage = "Examples of how to use this"
	d.HideVersion = true

	d.Add(cleanSlate(), "clean-slate", "Ensure a pristine demo environment")
	d.Add(installMetricsServer(), "setup-metrics", "Install the metrics server")
	d.Add(cpuDemoSetup(), "setup-cpu", "Setup the CPU workloads")
	d.Add(loadCPU(), "load-cpu", "Load workloads with CPU utilization")
	d.Add(cpuTop(), "cpu-top", "Show CPU utilization")
	d.Add(hpaDemoSetup(), "setup-hpa", "Setup the HPA")
	d.Add(showHPA(), "show-hpa", "Show the HPA")
	d.Add(cpuTopC2(), "cpu-top-c2", "Show the HPA C2")
	d.Add(cpuTopC3(), "cpu-top-c3", "Show the HPA C3")

	d.Run()
}

func cleanSlate() *demo.Run {
	r := demo.NewRun(
		`Clean Slate`,
	)

	r.Step(demo.S(
		`Delete any resources from previous demos`,
	), demo.S(
		`kubectl delete --ignore-not-found ns hpa-cpu-demo;`,
		`kubectl delete --ignore-not-found -n kube-system deployment metrics-server`,
	))

	return r
}

func installMetricsServer() *demo.Run {
	r := demo.NewRun(
		`Install the Metrics Server`,
	)

	r.Step(nil, demo.S(
		`kubectl apply -f lab0/metrics-server.yaml`,
	))

	r.Step(nil, demo.S(
		`kubectl get deployment metrics-server -n kube-system`,
	))

	return r
}

func cpuDemoSetup() *demo.Run {
	r := demo.NewRun(
		`Prepare the workload CPU demo resources`,
	)

	r.Step(demo.S(
		`Three identical workloads that consume CPU cycles but each with different CPU requests settings.`,
		`These will be used to demonstrate HPA on CPU.`,
		`Each workload will consume 450 millicores of CPU.`,
	), demo.S(
		`kubectl apply -k lab1/cpu/`,
	))

	r.Step(nil, demo.S(
		`kubectl get deployments -n hpa-cpu-demo`,
	))

	r.Step(demo.S(
		`Show the resources configuration for each pod.`,
		`Please note that "c1" does not have any CPU requests set.`,
		`The "c2" requests 500 millicores and "c3" requests only 50 millicores.`,
	), demo.S(
		`kubectl get pods -n hpa-cpu-demo -o yaml`,
		`| yq  '[ .items[] |`,
		`         {"name":   .metadata.name,`,
		`          "resources": {`,
		`            "requests":   {"cpu": .spec.containers[0].resources.requests.cpu  }}}]'`,
	))

	return r
}

func loadCPU() *demo.Run {
	r := demo.NewRun(
		`Inject CPU load into the workloads`,
	)

	r.Step(demo.S(
		`Load each workload with a 450m CPU of work.`,
	), demo.S(
		`kubectl port-forward svc/no-requests-no-limits 8081:8080 -n hpa-cpu-demo &`,
		`kubectl port-forward svc/requests-no-limits 8082:8080 -n hpa-cpu-demo &`,
		`kubectl port-forward svc/small-requests-no-limits 8083:8080 -n hpa-cpu-demo &`,
		`sleep 2;`,
		`curl --data "millicores=450&durationSec=3600" http://localhost:8081/ConsumeCPU;`,
		`curl --data "millicores=450&durationSec=3600" http://localhost:8082/ConsumeCPU;`,
		`curl --data "millicores=450&durationSec=3600" http://localhost:8083/ConsumeCPU;`,
		`kill %1 %2 %3; echo '...Done'`,
	))

	return r
}

func cpuTop() *demo.Run {
	r := demo.NewRun(
		`Inspect CPU usage of the pods`,
	)

	r.Step(nil, demo.S(
		`kubectl top pods -n hpa-cpu-demo`,
	))

	return r
}

func hpaDemoSetup() *demo.Run {
	r := demo.NewRun(
		`Create the HPA for the CPU workloads`,
	)

	r.Step(demo.S(
		`Create three identical HPA objects, one for each workload:`,
		`Every HPA to scale up to 5 replicas if CPU usage is above 50%`,
	), demo.S(
		`kubectl -n hpa-cpu-demo autoscale deployment c1-no-requests-no-limits --cpu-percent=50 --min=1 --max=5;`,
		`kubectl -n hpa-cpu-demo autoscale deployment c2-requests-no-limits --cpu-percent=50 --min=1 --max=5;`,
		`kubectl -n hpa-cpu-demo autoscale deployment c3-small-requests-no-limits --cpu-percent=50 --min=1 --max=5`,
	))

	/*
		r.Step(demo.S(
			`Show HPA objects:`,
		), demo.S(
			`kubectl -n hpa-cpu-demo get hpa`,
		))
	*/

	return r
}

func showHPA() *demo.Run {
	r := demo.NewRun(
		`Inspect HPA:`,
	)

	r.Step(nil, demo.S(
		`kubectl -n hpa-cpu-demo get hpa`,
	))

	return r
}

func cpuTopC2() *demo.Run {
	r := demo.NewRun(
		`Inspect CPU usage of the pods of big request workload`,
	)

	r.Step(nil, demo.S(
		`kubectl -n hpa-cpu-demo top pod -l name=requests-no-limits`,
	))

	r.Step(demo.S(
		`Note two replicas, but one is consuming the CPU, hence it is halving the 90% to 45%.`,
	), nil)

	return r
}

func cpuTopC3() *demo.Run {
	r := demo.NewRun(
		`Inspect CPU usage of the pods of small request workload`,
	)

	r.Step(nil, demo.S(
		`kubectl -n hpa-cpu-demo top pod -l name=small-requests-no-limits`,
	))

	r.Step(demo.S(
		`Note multiple replicas, but one is consuming most of the CPU.`,
		`HPA is accusing to 180% from 900% but it still high (requests is set too low).`,
		`Increasing the max replicas will make it worse (only one replica is busy).`,
	), nil)

	return r
}
