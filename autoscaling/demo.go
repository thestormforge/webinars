package main

import (
	demo "github.com/saschagrunert/demo"
)

const (
	pipeToColor string = "| colorize" // `| pygmentize -O style=material -l yaml || cat`
)

func main() {
	d := demo.New()

	d.Name = "Kubernetes autoscaling demo"
	d.Usage = "Examples of how to use this"
	d.HideVersion = true

	d.Add(cleanSlate(), "clean-slate", "Ensure a pristine demo environment")
	d.Add(installMetricsServer(), "setup-metrics", "Install the metrics server")
	d.Add(cpuDemoSetup(), "setup-cpu", "Setup the CPU demo")
	d.Add(cpuTop(), "cpu-top", "Setup the CPU demo")

	d.Run()
}

func cleanSlate() *demo.Run {
	r := demo.NewRun(
		`Clean Slate`,
	)

	r.Step(demo.S(
		`Delete any resources from previous demos`,
	), demo.S(
		`kubectl delete --ignore-not-found ns hpa-cpu-demo`,
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
