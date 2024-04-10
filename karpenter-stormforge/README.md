# Karpenter and StormForge

This repo is to support "Navigating Kubernetes Cost Optimization with Karpenter & StormForge"
webinar.

## Goal

- Show the cost comparison between Cluster Autoscaler and Karpenter
- Additional savings with StormForge setting the Requests/Limits

## Baseline Environment

- EKS cluster with a load generator running separately
- Sample application: https://github.com/GoogleCloudPlatform/microservices-demo with the exact same configuration set on three namespaces: sampleapp-on-CAS and sampleapp-on-karpenter and sampleapp-on-karpenter-optimized. We have custom helm charts to install this application along with a helm chart just for the load generator.
- `sampleapp-on-CAS` uses CAS over nodelabel=cluster-autoscaler
-- Show the CAS node configuration , we used m5.large (2 vCPU 8 GB RAM)
- `sampleapp-on-karpenter` uses Karpenter over nodelabel=karpenter with minimum size `.large` 
-- nodepool category: c, r , m and on-demand
-- Explain the rationale of the node sizes: we wanted a wide range to get karpenter to pickup the best ones and spot to maximize savings.
- `sampleapp-on-karpenter-optimized` uses Karpenter over nodelabel=karpenter-optimized with spot and compression mode 


## Demo 1: difference between CAS and Karpenter

- Cooking show 1:
- Have load generator driver insane amount of traffic to both applications
-- Will need to show how the load generator is configured (traffic 24x7, every 10 min spread different users, with random multiplier)
- Both nodegroups will scale
- Run eks-node-viewer on both of them and show the difference how they will scale

## Demo 2: difference on application on Karpenter with Compression mode

- See the price difference

## Demo 3: Cost savings after StormForge patches

- Load StormForge UI and install StormForge agent
- Cooking show 2:
- The cluster with all recommendations ready to be applied, after 7 days 
- Show the patch and apply it automatically
- Show how small the footprint is

## Conclusion

- Karpenter is much more cost effective than CAS by itself. Without `spot` instances we see 50% savings.
- Karpenter with `spot` instances plus compression mode gives extra gains.
- However, Karpenter still depends on Kubernetes Requests, StormForge will configure it for you. The `frontend` microservice needed more CPU but all other micro-services needed much less resources.

## Provision Environment

Add your StormForge credentials into `eks.tf` . The credentials can be obtained signing up to StormForge here: https://app.stormforge.io/signup

```hcl
stormforge = {
      name = "stormforge-agent"
      description = "StormForge agent"
      repository = "oci://registry.stormforge.io/library/"
      chart = "stormforge-agent"
      create_namespace = true
      namespace = "stormforge-system"
      values = [
        <<-EOT
          clusterName: ${module.eks.cluster_name}
          stormforge:
            address: https://api.stormforge.io/
          authorization:
            issuer: https://api.stormforge.io/
            clientID: ADD YOUR CLIENT ID HERE
            clientSecret: ADD YOUR CLIENT SECRET HERE
        EOT
      ]
    }
```

Change other desired values in `locals` of `main.tf` file:

```hcl
locals {
  name   = "stormforge-demo"
  region = "us-west-2" # Change to your desired region

  vpc_cidr = "10.1.0.0/16"
  azs      = slice(data.aws_availability_zones.available.names, 0, 3)

  tags = {
    Blueprint  = local.name
    GithubRepo = "github.com/aws-ia/terraform-aws-eks-blueprints"
  }
}
```

Apply terraform:

```bash
terraform init && terraform apply --auto-approve
```

## Webinars Given

### March-2024

[YouTube Short, Rafa Brito and Lucas Duarte](https://www.youtube.com/watch?v=RbOg0aZyQTw)
