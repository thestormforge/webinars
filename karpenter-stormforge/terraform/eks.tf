################################################################################
# Cluster
################################################################################

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.21"

  cluster_name                   = local.name
  cluster_version                = local.cluster_version
  cluster_endpoint_public_access = true

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  # Fargate profiles use the cluster primary security group so these are not utilized
  create_cluster_security_group = false
  create_node_security_group    = false

  manage_aws_auth_configmap = true
  aws_auth_roles = [
    # We need to add in the Karpenter node IAM role for nodes launched by Karpenter
    {
      rolearn  = module.eks_blueprints_addons.karpenter.node_iam_role_arn
      username = "system:node:{{EC2PrivateDNSName}}"
      groups = [
        "system:bootstrappers",
        "system:nodes",
      ]
    },
  ]

  eks_managed_node_groups = {
    cluster_autoscaler = {
      instance_types = ["m5.large"]
      min_size       = 1
      max_size       = 30
      desired_size   = 1

      labels = {
        provisionerType = "cluster-autoscaler"
        GithubRepo      = "example.git.com"
      }

    }

    infra_node_group = {
      instance_types = ["m5.large"]
      min_size       = 3
      max_size       = 5
      desired_size   = 3

      labels = {
        provisionerType = "infra"
        GithubRepo      = "example.git.com"
      }

    }
  }
  tags = merge(local.tags, {
    # NOTE - if creating multiple security groups with this module, only tag the
    # security group that Karpenter should utilize with the following tag
    # (i.e. - at most, only one security group should have this tag in your account)
    "karpenter.sh/discovery" = local.name
  })
}

################################################################################
# EKS Blueprints Addons
################################################################################

module "eks_blueprints_addons" {
  source  = "aws-ia/eks-blueprints-addons/aws"
  version = "~> 1.2"

  cluster_name      = module.eks.cluster_name
  cluster_endpoint  = module.eks.cluster_endpoint
  cluster_version   = module.eks.cluster_version
  oidc_provider_arn = module.eks.oidc_provider_arn

  # We want to wait for the Fargate profiles to be deployed first
  create_delay_dependencies = [for prof in module.eks.fargate_profiles : prof.fargate_profile_arn]

  eks_addons = {
    coredns    = {}
    vpc-cni    = {}
    kube-proxy = {}
  }
  enable_metrics_server = true

  enable_karpenter = true
  karpenter = {
    repository_username = data.aws_ecrpublic_authorization_token.token.user_name
    repository_password = data.aws_ecrpublic_authorization_token.token.password
  }
  karpenter_node = {
    # Use static name so that it matches what is defined in `karpenter.yaml` example manifest
    iam_role_use_name_prefix = false
  }

  enable_cluster_autoscaler = true

  helm_releases = {
    stormforge = {
      name             = "stormforge-agent"
      description      = "StormForge agent"
      repository       = "oci://registry.stormforge.io/library/"
      chart            = "stormforge-agent"
      create_namespace = true
      namespace        = "stormforge-system"
      values = [
        <<-EOT
          clusterName: ${module.eks.cluster_name}
          stormforge:
            address: https://api.stormforge.io/
          authorization:
            issuer: https://api.stormforge.io/
            clientID: ${var.stormforge_client_id}
            clientSecret: ${var.stormforge_client_secret}
          clusterData: true
        EOT
      ]
    }

    stormforge-applier = {
      name             = "stormforge-applier"
      description      = "StormForge applier"
      repository       = "oci://registry.stormforge.io/library/"
      chart            = "stormforge-applier"
      create_namespace = true
      namespace        = "stormforge-system"
      values = [
        <<-EOT
          clusterName: ${module.eks.cluster_name}
          authorization:
            issuer: https://api.stormforge.io/
            clientID: ${var.stormforge_client_id}
            clientSecret: ${var.stormforge_client_secret}
        EOT
      ]
    }

    stormforge-loadgen-karpenter = {
      name             = "stormforge-loadgen-karpenter"
      description      = "StormForge Load Gen Hipster App"
      repository       = "https://registry.stormforge.io/chartrepo/examples"
      chart            = "sf-hipster-shop-loadgenerator"
      create_namespace = true
      namespace        = "sampleapp-on-karpenter"
      values = [
        "${file("./helm-values/load-gen.yaml")}"
      ]
    }

    stormforge-loadgen-cas = {
      name             = "stormforge-loadgen-cas"
      description      = "StormForge Load Gen Hipster App"
      repository       = "https://registry.stormforge.io/chartrepo/examples"
      chart            = "sf-hipster-shop-loadgenerator"
      create_namespace = true
      namespace        = "sampleapp-on-cas"
      values = [
        "${file("./helm-values/load-gen.yaml")}"
      ]
    }

    stormforge-hipsterapp-karpenter = {
      name             = "stormforge-hipsterapp-karpenter"
      description      = "StormForge Hipster App"
      repository       = "https://registry.stormforge.io/chartrepo/examples"
      chart            = "sf-hipster-shop"
      create_namespace = true
      namespace        = "sampleapp-on-karpenter"
      values = [
        "${file("./helm-values/inflate-app-karpenter.yaml")}"
      ]
    }

    stormforge-hipsterapp-cas = {
      name             = "stormforge-hipsterapp-cas"
      description      = "StormForge Hipster App"
      repository       = "https://registry.stormforge.io/chartrepo/examples"
      chart            = "sf-hipster-shop"
      create_namespace = true
      namespace        = "sampleapp-on-cas"
      values = [
        "${file("./helm-values/inflate-app-cas.yaml")}"
      ]
    }

    stormforge-loadgen-karpenter-optimized = {
      name             = "stormforge-loadgen-karpenter-optimized"
      description      = "StormForge Load Gen Hipster App"
      repository       = "https://registry.stormforge.io/chartrepo/examples"
      chart            = "sf-hipster-shop-loadgenerator"
      create_namespace = true
      namespace        = "sampleapp-on-karpenter-optimized"
      values = [
        "${file("./helm-values/load-gen.yaml")}"
      ]
    }

    stormforge-hipsterapp-karpenter-optimized = {
      name             = "stormforge-hipsterapp-karpenter"
      description      = "StormForge Hipster App"
      repository       = "https://registry.stormforge.io/chartrepo/examples"
      chart            = "sf-hipster-shop"
      create_namespace = true
      namespace        = "sampleapp-on-karpenter-optimized"
      values = [
        "${file("./helm-values/inflate-app-karpenter-optimized.yaml")}"
      ]
    }

  }

  tags = local.tags
}

module "eks_data_addons" {
  source = "aws-ia/eks-data-addons/aws"
  version = "~> 1.31.3" # ensure to update this to the latest/desired version

  oidc_provider_arn = module.eks.oidc_provider_arn
  #---------------------------------------
  # Deploying Karpenter resources(Nodepool and NodeClass) with Helm Chart
  #---------------------------------------
  enable_karpenter_resources = true
  karpenter_resources_helm_config = {
    microsservice-demo = {
      values = [
        <<-EOT
      name: microsservice-demo
      clusterName: ${module.eks.cluster_name}
      ec2NodeClass:
        karpenterRole: ${split("/", module.eks_blueprints_addons.karpenter.node_iam_role_arn)[1]}
        subnetSelectorTerms:
          id: ${module.vpc.private_subnets[2]}
        securityGroupSelectorTerms:
          tags:
            karpenter.sh/discovery: ${module.eks.cluster_name}
        blockDevice:
          deviceName: /dev/xvda
          volumeSize: 100Gi
          volumeType: gp3
          encrypted: true
          deleteOnTermination: true
      nodePool:
        labels:
          - provisionerType: karpenter
        requirements:
          - key: "karpenter.k8s.aws/instance-category"
            operator: In
            values: ["c", "m", "r"]
          - key: "karpenter.k8s.aws/instance-size"
            operator: In
            values: ["large", "xlarge", "2xlarge", "4xlarge", "8xlarge", "16xlarge", "24xlarge"]
          - key: "kubernetes.io/arch"
            operator: In
            values: ["amd64"]
          - key: "karpenter.sh/capacity-type"
            operator: In
            values: ["on-demand"]
        limits:
          cpu: 1000
        disruption:
          consolidationPolicy: WhenEmpty
          consolidateAfter: 30s
          expireAfter: 720h
        weight: 100
      EOT
      ]
    }

    microsservice-demo-optimized = {
      values = [
        <<-EOT
      name: microsservice-demo-optimized
      clusterName: ${module.eks.cluster_name}
      ec2NodeClass:
        karpenterRole: ${split("/", module.eks_blueprints_addons.karpenter.node_iam_role_arn)[1]}
        subnetSelectorTerms:
          id: ${module.vpc.private_subnets[2]}
        securityGroupSelectorTerms:
          tags:
            karpenter.sh/discovery: ${module.eks.cluster_name}
        blockDevice:
          deviceName: /dev/xvda
          volumeSize: 100Gi
          volumeType: gp3
          encrypted: true
          deleteOnTermination: true
      nodePool:
        labels:
          - provisionerType: karpenter-optimized
        requirements:
          - key: "karpenter.k8s.aws/instance-category"
            operator: In
            values: ["c", "m", "r"]
          - key: "karpenter.k8s.aws/instance-size"
            operator: In
            values: ["large", "xlarge", "2xlarge", "4xlarge", "8xlarge", "16xlarge", "24xlarge"]
          - key: "kubernetes.io/arch"
            operator: In
            values: ["amd64"]
          - key: "karpenter.sh/capacity-type"
            operator: In
            values: ["spot"]
        limits:
          cpu: 1000
        disruption:
          consolidationPolicy: WhenUnderutilized
          expireAfter: 720h
        weight: 100
      EOT
      ]
    }
  }
}