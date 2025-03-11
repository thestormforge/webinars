locals {
  region                   = "us-east-1"
  vpc_cidr                 = "10.0.0.0/16"
  name                     = "eks-automode-cluster-demo-2"
  cluster_version          = "1.32"
  azs                      = slice(data.aws_availability_zones.available.names, 0, 3)
  stormforge_client_id     = var.stormforge_client_id
  stormforge_client_secret = var.stormforge_client_secret
  enable_applier           = true

  tags = {
    Blueprint = local.name
  }
}

variable "stormforge_client_id" {
  default = ""
}

variable "stormforge_client_secret" {
  default = "" 
}


# Define the required providers
provider "aws" {
  region = local.region # Change to your desired region
}

provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    # This requires the awscli to be installed locally where Terraform is executed
    args = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
  }
}

provider "helm" {
  kubernetes {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)

    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      # This requires the awscli to be installed locally where Terraform is executed
      args = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
    }
  }
}

data "aws_availability_zones" "available" {
  # Do not include local zones
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

# Use the Terraform VPC module to create a VPC
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.17.0" # Use the latest version available

  name = "${local.name}-vpc"
  cidr = local.vpc_cidr

  azs             = local.azs
  private_subnets = [for k, v in local.azs : cidrsubnet(local.vpc_cidr, 4, k)]
  public_subnets  = [for k, v in local.azs : cidrsubnet(local.vpc_cidr, 8, k + 48)]

  enable_nat_gateway = true
  single_nat_gateway = true

  public_subnet_tags = {
    "kubernetes.io/role/elb" = 1
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = 1
  }

  tags = local.tags
}

# Use the Terraform EKS module to create an EKS cluster
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "20.33.1" # Use the latest version available

  cluster_name    = local.name
  cluster_version = local.cluster_version

  cluster_endpoint_public_access           = true
  enable_irsa                              = true
  enable_cluster_creator_admin_permissions = true

  cluster_addons = {
    metrics-server = {}
  }

  cluster_compute_config = {
    enabled    = true
    node_pools = ["general-purpose", "system"]
  }


  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  tags = local.tags
}


resource "helm_release" "stormforge_agent" {
  name             = "stormforge-agent"
  description      = "StormForge agent"
  repository       = "oci://registry.stormforge.io/library/"
  chart            = "stormforge-agent"
  namespace        = "stormforge-system"
  create_namespace = true

  values = [
    <<-EOT
      clusterName: ${module.eks.cluster_name}
      stormforge:
        address: https://api.stormforge.io/
      authorization:
        issuer: https://api.stormforge.io/
        clientID: ${local.stormforge_client_id}
        clientSecret: ${local.stormforge_client_secret}
      clusterData: true
    EOT
  ]

  # Optional: Add timeout
  timeout = 1200 # 20 minutes

  # Optional: Add dependencies
  depends_on = [
    module.eks
  ]
}

resource "helm_release" "stormforge_applier" {
  count            = local.enable_applier ? 1 : 0
  name             = "stormforge-applier"
  description      = "StormForge applier"
  repository       = "oci://registry.stormforge.io/library/"
  chart            = "stormforge-applier"
  namespace        = "stormforge-system"
  create_namespace = true

  values = [
    <<-EOT
      clusterName: ${module.eks.cluster_name}
      authorization:
        issuer: https://api.stormforge.io/
        clientID: ${local.stormforge_client_id}
        clientSecret: ${local.stormforge_client_secret}
    EOT
  ]

  depends_on = [
    module.eks, helm_release.stormforge_agent
  ]
}



# Hipster App for optimization
resource "helm_release" "stormforge_hipster" {
  name             = "stormforge-hipsterapp-automode"
  description      = "StormForge Hipster App"
  repository       = "oci://registry.stormforge.io/examples/"
  chart            = "sf-hipster-shop"
  namespace        = "sampleapp-on-automode"
  create_namespace = true

  values = [
    file("./helm-values/inflate-app-automode.yaml")
  ]

  # Optional: Add timeout
  timeout = 1800 # 30 minutes

  # Optional: Add dependencies if needed
  depends_on = [
    module.eks # Assuming you're using the EKS module from your open file
  ]
}

resource "helm_release" "stormforge_loadgen_automode" {
  name             = "stormforge-loadgen-automode"
  description      = "StormForge Load Gen Hipster App"
  repository       = "oci://registry.stormforge.io/examples/"
  chart            = "sf-hipster-shop-loadgenerator"
  namespace        = "sampleapp-on-automode"
  create_namespace = true

  values = [
    file("./helm-values/load-gen.yaml")
  ]

  timeout = 1800 # 30 minutes

  depends_on = [module.eks, helm_release.stormforge_hipster]
}



# Outputs
output "configure_kubectl" {
  description = "Configure kubectl: make sure you're logged in with the correct AWS profile and run the following command to update your kubeconfig"
  value       = "aws eks --region ${local.region} update-kubeconfig --name ${module.eks.cluster_name}"
}

