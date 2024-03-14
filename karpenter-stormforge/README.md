# Karpenter and StormForge

This repo is to support "Navigating Kubernetes Cost Optimization with Karpenter & StormForge"
webinar.

## Goal

- Show the cost comparison between Cluster Autoscaler and Karpenter
- Additional savings with StormForge setting the Requests/Limits

## Baseline Environment

<Perhaps a mermaid diagram here>
- (Managed?) EKS cluster with a load generator running separatly
- [Sample application][https://github.com/GoogleCloudPlatform/microservices-demo] with the exact same configuration set on two namespaces: app-on-CAS and app-on-Karpenter
- `app-on-CAS` uses CAS over nodelabel=CAS
-- Show the CAS node configguration , we used m5.large (2 vCPU 8 GB RAM)
- `app-on-Karpenter` uses Karpenter over nodelabel=karpenter (what is the nodepool configuration)
-- nodepool category: c, r , m. spot and on-demand
-- Explain the rationale of the node sizes: we wanted a wide range to get karpenter to pickup the best ones and spot to maximize savings.


## Demo 1: difference between CAS and Karpenter

- Cooking show 1?
- Have load generator driver insane amount of traffic to both applications
-- Will need to show how the load generator is configured (traffic 24x7 ?, during biz hours, amount of traffic, etc?)
- Both nodegroups will scale
- Run eks-node-viewer on both of them and show the difference how they will scale

## Demo 2: difference on application on Karpenter after StormForge

- Load StormForge UI and install StormForge agent
- Cooking show 2?
- The cluster with all recommendations ready to be applied, after 7 days 
- Show the patch and apply it automatically
- Show how small the footprint is

## Conclusion

- Karpenter is much better than CAS
- Bit Karpenter still depends on Kubernetes Requests, StormForge will configure it for you
