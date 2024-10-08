package main

import (
	demo "github.com/saschagrunert/demo"
)

const (
	pipeToColor string = "| colorize" // `| pygmentize -O style=material -l yaml || cat`
)

func main() {
	d := demo.New()

	d.Name = "Kubernetes resources demo"
	d.Usage = "Examples of how to use this"
	d.HideVersion = true

	d.Add(cleanSlate(), "clean-slate", "Ensure a pristine demo environment")
	d.Add(inspectNodes(), "inspect-nodes", "Inspect the demo cluster nodes")
	d.Add(cpuDemoSetup(), "setup-cpu", "Setup the CPU demo")
	d.Add(cpuDemo(), "cpu", "Show a series of CPU load demonstrations")
	d.Add(cpuDemoRequests(), "cpu-requests", "Show CPU load demonstrations on Requests")
	d.Add(cpuDelete(), "cpu-delete", "Delete the CPU demo resources")
	d.Add(cpuDemoLimits(), "cpu-limits", "Show CPU load demonstrations on Limits")
	d.Add(memoryDemoSetup(), "setup-memory", "Setup the Memory demo")
	d.Add(memoryDemoApplicationOOM(), "memory-app-oom", "Show Application OOM event")
	d.Add(memoryDemoSystemOOM(), "memory-system-oom", "Show System OOM events")
	d.Add(memoryDemoEviction(), "memory-eviction", "Show Eviction events")
	d.Add(memoryDelete(), "memory-delete", "Delete the memory demo resources")
	d.Add(setupPoliciesDemo(), "setup-policies", "Setup the policies demo")
	d.Add(limitRangesDemo(), "limit-ranges", "Demonstrate LimitRanges")
	d.Add(quotaDemo(), "quotas", "Demonstrate ResourceQuotas")

	d.Run()
}

func cleanSlate() *demo.Run {
	r := demo.NewRun(
		`Clean Slate`,
	)

	r.Step(demo.S(
		`Delete any resources that might be lingering around from the last demo to ensure a pristine environment`,
	), demo.S(
		`kubectl delete --ignore-not-found -k .`,
	))

	r.Step(demo.S(
		`Create the monitoring stack for the cluster`,
	), demo.S(
		`kubectl apply -k lab0/monitoring/`,
	))

	return r
}

func inspectNodes() *demo.Run {
	r := demo.NewRun(
		`Inspect the demo cluster`,
		`Show that the cluster has only one worker node, what its available resources are,`,
		`and which workloads/pods are running to start with.`,
	)

	r.Step(demo.S(
		`Node details for the cluster`,
	), demo.S(
		`kubectl get nodes -l node-role.kubernetes.io/control-plane!=""`,
	))

	r.Step(nil, demo.S(
		`kubectl get nodes -l node-role.kubernetes.io/control-plane!="" -o yaml`,
		`| yq '.items[] | `,
		`        {.metadata.name:`,
		`          {"capacity":    .status.capacity,`,
		`           "allocatable": .status.allocatable}}'`,
	))

	r.Step(nil, demo.S(
		`kubectl top nodes -l node-role.kubernetes.io/control-plane!=""`,
	))

	return r
}

func cpuDemoSetup() *demo.Run {
	r := demo.NewRun(
		`Prepare the CPU demo resources`,
	)

	r.Step(demo.S(
		`Create CPU demo workloads`,
		`These will be used to demonstrate requests, limits, and effects of resource`,
		`contention for CPU.`,
	), demo.S(
		`kubectl apply -k lab1/cpu/`,
	))

	r.Step(demo.S(
		`The demo workloads`,
	), demo.S(
		`kubectl get deployments -n cpu`,
	))

	r.Step(nil, demo.S(
		`kubectl get pods -o wide -n cpu`,
	))

	r.Step(demo.S(
		`Show the resources configuration for each pod.`,
	), demo.S(
		`kubectl get pods -n cpu -o yaml`,
		`| yq  '[ .items[] |`,
		`         {"name":   .metadata.name,`,
		`          "qosClass": .status.qosClass,`,
		`          "resources": {`,
		`            "requests": {"cpu": .spec.containers[0].resources.requests.cpu},`,
		`            "limits":   {"cpu": .spec.containers[0].resources.limits.cpu  }}}]'`,
	))

	return r
}

func cpuDemo() *demo.Run {
	r := demo.NewRun(
		`Baseline for a series of CPU loading demonstrations`,
	)

	r.Step(demo.S(
		`Load each workload with a BASELINE 450m CPU of work.`,
		``,
		`    no-requests-no-limits: 450m`,
		`    requests-no-limits:    450m`,
		`    requests-and-limits:   450m`,
		``,
	), demo.S(
		`kubectl port-forward svc/no-requests-no-limits 8081:8080 -n cpu &`,
		`kubectl port-forward svc/requests-no-limits 8082:8080 -n cpu &`,
		`kubectl port-forward svc/requests-and-limits 8083:8080 -n cpu &`,
		`sleep 2;`,
		`curl --data "millicores=450&durationSec=3600" http://localhost:8081/ConsumeCPU;`,
		`curl --data "millicores=450&durationSec=3600" http://localhost:8082/ConsumeCPU;`,
		`curl --data "millicores=450&durationSec=3600" http://localhost:8083/ConsumeCPU;`,
		`kill %1 %2 %3; echo '...Done'`,
	))

	r.Step(demo.S(
		`On Grafana (in a separate window)`,
		``,
		`    three workload pods, each consuming 450m CPU`,
	), nil)

	return r
}

func cpuDemoRequests() *demo.Run {
	r := demo.NewRun(
		`CPU loading demonstrations for requests`,
	)

	r.Step(demo.S(
		`Make requests-no-limits try and overconsume an ADDITIONAL 2000m of CPU for 15s.`,
		``,
		`    no-requests-no-limits: 450m`,
		`    requests-no-limits:    2000m`,
		`    requests-and-limits:   450m`,
		``,
	), demo.S(
		`kubectl port-forward svc/requests-no-limits 8082:8080 -n cpu &`,
		`sleep 2;`,
		`curl --data "millicores=1000&durationSec=60" http://localhost:8082/ConsumeCPU;`,
		`curl --data "millicores=1000&durationSec=60" http://localhost:8082/ConsumeCPU;`,
		`kill %1; echo '...Done'`,
	))

	r.Step(demo.S(
		`What do you see on Grafana (in a separate window)?`,
	), nil)

	return r
}

func cpuDemoLimits() *demo.Run {
	r := demo.NewRun(
		`CPU loading demonstrations for limits`,
	)

	r.Step(demo.S(
		`Repeat, but with requests-and-limits trying to overconsume this time.`,
		``,
		`    no-requests-no-limits: 450m`,
		`    requests-no-limits:    450m`,
		`    requests-and-limits:   2000m`,
		``,
	), demo.S(
		`kubectl port-forward svc/requests-and-limits 8083:8080 -n cpu &`,
		`sleep 2;`,
		`curl --data "millicores=1000&durationSec=30" http://localhost:8083/ConsumeCPU;`,
		`curl --data "millicores=1000&durationSec=30" http://localhost:8083/ConsumeCPU;`,
		`kill %1; echo '...Done'`,
	))

	r.Step(demo.S(
		`What do you see on Grafana (in a separate window)?`,
	), nil)

	return r
}

func cpuDelete() *demo.Run {
	r := demo.NewRun(
		`Delete namespace of CPU loading demonstrations`,
	)

	r.Step(demo.S(
		`Delete the demo namespace and resources.`,
	), demo.S(
		`kubectl delete --ignore-not-found -k lab1/cpu/`,
	))

	return r
}

func memoryDemoSetup() *demo.Run {
	r := demo.NewRun(
		`Prepare the Memory demo resources`,
	)

	r.Step(demo.S(
		`Create Memory demo workloads`,
		`These will be used to demonstrate requests, limits, and effects of resource`,
		`contention for Memory.`,
	), demo.S(
		`kubectl apply -k lab2/memory/`,
	))

	r.Step(demo.S(
		`The demo workloads`,
	), demo.S(
		`kubectl get deployments -n memory`,
	))

	r.Step(nil, demo.S(
		`kubectl get pods -o wide -n memory`,
	))

	r.Step(demo.S(
		`Show the resources configuration for each pod.`,
	), demo.S(
		`kubectl get pods -n memory -o yaml`,
		`| yq  '[ .items[] |`,
		`         {"name":   .metadata.name,`,
		`          "qosClass": .status.qosClass,`,
		`          "resources": {`,
		`            "requests": .spec.containers[0].resources.requests.memory,`,
		`            "limits":   .spec.containers[0].resources.limits.memory}} ]'`,
	))

	r.Step(demo.S(
		`Each workload has a BASELINE 150 MiB Memory.`,
		``,
		`    no-requests-no-limits:  ~150 MiB`,
		`    requests-no-limits:     ~150 MiB`,
		`    requests-and-limits:    ~150 MiB`,
		``,
	), nil)

	r.Step(demo.S(
		`You should see on Grafana each workload running and consuming ~150 MiB?`,
	), nil)

	return r
}

func memoryDemoApplicationOOM() *demo.Run {
	r := demo.NewRun(
		`Perform a series of Memory loading demonstrations - Application OOM`,
	)

	r.Step(demo.S(
		`Try loading requests-and-limits with an additional 500 MiB.`,
		``,
		`    no-requests-no-limits:  ~150 MiB`,
		`    requests-no-limits:     ~150 MiB`,
		`    requests-and-limits:    ~500 GiB ^`,
		``,
	), demo.S(
		`kubectl port-forward svc/requests-and-limits 8083:8080 -n memory &`,
		`sleep 2;`,
		`curl --data '{"mebibytes": 500, "seconds": 30, "delay": 1}' http://localhost:8083/ConsumeMem;`,
		`kill %1; echo '...Done'`,
	))

	r.Step(demo.S(
		`Inspect the requests-and-limits pod.`,
	), demo.S(
		`kubectl get pods -l name=requests-and-limits -n memory`,
	))

	r.Step(nil, demo.S(
		`kubectl describe pod -l name=requests-and-limits -n memory | egrep -A 21 '^Containers:$'`,
	))

	r.Step(demo.S(
		`Inspect again the requests-and-limits pod.`,
	), demo.S(
		`kubectl get pods -l name=requests-and-limits -n memory`,
	))

	return r
}

func memoryDemoSystemOOM() *demo.Run {
	r := demo.NewRun(
		`Perform a series of Memory loading demonstrations - System OOM`,
	)

	r.Step(demo.S(
		`Try loading no-requests-no-limits and requests-no-limits with an additional 250 MiB.`,
		``,
		`    no-requests-no-limits:  ~400 MiB ^`,
		`    requests-no-limits:     ~400 MiB ^`,
		`    requests-and-limits:    ~150 MiB (Pod will restart after OOM)`,
		``,
	), demo.S(
		`kubectl port-forward svc/no-requests-no-limits 8081:8080 -n memory &`,
		`kubectl port-forward svc/requests-no-limits 8082:8080 -n memory &`,
		`sleep 2;`,
		`curl --data '{"mebibytes": 250, "seconds": 600, "delay": 1}' http://localhost:8081/ConsumeMem;`,
		`curl --data '{"mebibytes": 250, "seconds": 600, "delay": 1}' http://localhost:8082/ConsumeMem;`,
		`kill %1 %2; echo '...Done'`,
	))

	r.Step(demo.S(
		`What do you see on Grafana?`,
		`What happened to the no-requests-no-limits pod?`,
	), nil)

	r.Step(nil, demo.S(
		`kubectl describe pod -l name=no-requests-no-limits -n memory | egrep -A 21 '^Containers:$'`,
	))

	r.Step(demo.S(
		`What happened was the linux kernel OOM killed the process.`,
		`The pod used more memory than the node had available.`,
		`System OOM acted before the pod was evicted by kubelet.`,
		`This process was picked because it had QoS class BestEffort.`,
		``,
	), demo.S(
		`kubectl get events | grep OOM | tail -1`,
	))

	return r
}

func memoryDemoEviction() *demo.Run {
	r := demo.NewRun(
		`Perform a series of Memory loading demonstrations - Eviction`,
		`This is just tricky to generate, we need to avoid the system oom to act first.`,
	)

	r.Step(nil, demo.S(
		`kubectl top nodes -l node-role.kubernetes.io/control-plane!=""`,
	))

	r.Step(nil, demo.S(
		`kubectl top pods -n memory`,
	))

	r.Step(demo.S(
		`Try loading no-requests-no-limits with an additional 150 MiB, 50 MiB at time.`,
		``,
		`Node has a hard eviction threshold of 200 MiB.`,
		`    no-requests-no-limits:  ~300 MiB ^ incrementally increasing until eviction`,
		`    requests-no-limits:     ~400 MiB`,
		`    requests-and-limits:    ~150 MiB`,
		``,
	), demo.S(
		`kubectl port-forward svc/no-requests-no-limits 8081:8080 -n memory &`,
		`sleep 2;`,
		`curl --data '{"mebibytes": 50, "seconds": 600, "delay": 1}' http://localhost:8081/ConsumeMem;`,
		`sleep 1;`,
		`curl --data '{"mebibytes": 50, "seconds": 600, "delay": 1}' http://localhost:8081/ConsumeMem;`,
		`sleep 1;`,
		`curl --data '{"mebibytes": 50, "seconds": 600, "delay": 1}' http://localhost:8081/ConsumeMem;`,
		`sleep 1;`,
		`kill %1; echo '...Done'`,
	))

	r.Step(nil, demo.S(
		`kubectl get pod -l name=no-requests-no-limits -n memory`,
	))

	return r
}

func memoryDelete() *demo.Run {
	r := demo.NewRun(
		`Delete Memory workload loading demonstrations`,
	)

	r.Step(demo.S(
		`Delete the demo namespace and resources.`,
	), demo.S(
		`kubectl delete --ignore-not-found -k lab2/memory/`,
	))

	return r
}

func setupPoliciesDemo() *demo.Run {
	r := demo.NewRun(
		`Prepare the policy demo resources`,
	)

	r.Step(demo.S(
		`Clear and create the namespace`,
	), demo.S(
		`kubectl delete --ignore-not-found -k policies/;`,
		`kubectl create -f policies/namespace.yaml`,
	))

	return r
}

func limitRangesDemo() *demo.Run {
	r := demo.NewRun(
		`Show how LimitRanges work`,
	)

	r.Step(demo.S(
		`Show the LimitRange definition`,
	), demo.S(
		`cat policies/default-resources.yaml`,
		pipeToColor,
	))

	r.Step(demo.S(
		`Show the resource requests and limits defined in a basic deployment.`,
	), demo.S(
		`cat policies/no-requests.yaml`,
		pipeToColor,
	))

	r.Step(demo.S(
		`Create the basic deployment in the policies namespace.`,
	), demo.S(
		`kubectl apply -f policies/no-requests.yaml`,
	))

	r.Step(demo.S(
		`Now apply the LimitRange policy.`,
	), demo.S(
		`kubectl apply -f policies/default-resources.yaml`,
	))

	r.Step(demo.S(
		`What are the running pod's resource settings?`,
	), demo.S(
		`kubectl get pod -l name=no-requests -n policies -o yaml`,
		`| yq '.items[] | .spec.containers[] | {"resources": .resources}'`,
		pipeToColor,
	))

	r.Step(demo.S(
		`Delete the pod, forcing it to be recreated.`,
	), demo.S(
		`kubectl delete pod -l name=no-requests -n policies`,
	))

	r.Step(demo.S(
		`What are the resource settings now?`,
	), demo.S(
		`kubectl get pod -l name=no-requests -n policies -o yaml`,
		`| yq '.items[] | .spec.containers[] | {"resources": .resources}'`,
		pipeToColor,
	))

	r.Step(demo.S(
		`Note that LimitRanges work on pod admission; they don't modify deployments.`,
	), demo.S(
		`kubectl get deployment no-requests -n policies -o yaml`,
		`| yq '{"spec": {"template": .spec.template}}'`,
		pipeToColor,
	))

	r.Step(demo.S(
		`Clean up the policy and deployment.`,
	), demo.S(
		`kubectl delete limitranges --all -n policies;`,
		`kubectl delete -f policies/no-requests.yaml -n policies;`,
	))

	return r
}

func quotaDemo() *demo.Run {
	r := demo.NewRun(
		`Show how ResourceQuotas work`,
	)

	r.Step(demo.S(
		`Show the ResourceQuota definition`,
	), demo.S(
		`cat policies/default-resources-quota.yaml`,
		pipeToColor,
	))

	r.Step(demo.S(
		`Show the resource requests and limits defined in a basic deployment.`,
	), demo.S(
		`cat policies/test-workload.yaml`,
		pipeToColor,
	))

	r.Step(demo.S(
		`Create the basic deployment in the policies namespace.`,
	), demo.S(
		`kubectl apply -f policies/test-workload.yaml`,
	))

	r.Step(demo.S(
		`Show the current requests of all containers in the namespace`,
	), demo.S(
		`kubectl get pods -n policies -o yaml`,
		`| yq '[ .items[] |`,
		`        {"name": .metadata.name,`,
		`         "requests": .spec.containers[].resources.requests} ]'`,
		pipeToColor,
	))

	r.Step(demo.S(
		`Apply the ResourceQuota to the policies namespace.`,
	), demo.S(
		`kubectl apply -f policies/default-resources-quota.yaml -n policies`,
	))

	r.Step(demo.S(
		`Show the workload resources for the policies namespace.`,
	), demo.S(
		`kubectl get deployments,replicasets -n policies; echo;`,
		`kubectl get pods -o custom-columns-file=policies/pod-custom-columns.txt -n policies`,
	))

	r.Step(demo.S(
		`Try scaling the workload up from 2 replicas to 4.`,
	), demo.S(
		`kubectl scale deployment test-workload --replicas=4 -n policies`,
	))

	r.Step(demo.S(
		`How did that go?`,
	), demo.S(
		`kubectl get deployments,replicasets -n policies; echo;`,
		`kubectl get pods -o custom-columns-file=policies/pod-custom-columns.txt -n policies`,
	))

	r.Step(demo.S(
		`Show the quota-related events on the current workload ReplicaSet`,
	), demo.S(
		`kubectl describe -n policies $(kubectl get rs -l name=test-workload -n policies -o name | head -1)`,
		`| sed -n '/Events:/, /Name:/ p'`,
	))

	r.Step(demo.S(
		`What about actual usage?`,
		``,
		`    kubectl port-forward -n monitoring svc/grafana 3000:3000`,
		``,
		`URL:`,
		``,
		`  http://localhost:3000/d/resource-usage-observatory`,
		`  Username: admin`,
		`  Password: admin`,
		``,
	), nil)

	r.Step(demo.S(
		`cleanup`,
	), demo.S(
		`kubectl delete resourcequotas --all -n policies;`,
		`kubectl delete -f policies/test-workload.yaml -n policies;`,
	))

	return r
}

/*
func testCommand() *demo.Run {
	r := demo.NewRun(
		`Test a command`,
	)

	r.Step(demo.S(
		`Test command`,
	), demo.S(
		`kubectl get pods`,
	))

	return r
}
*/
