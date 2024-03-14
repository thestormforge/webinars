```
% eksctl create cluster --name=webinar-20240314 --fargate --node-type=m5.large --nodes-min=1 --dry-run
accessConfig:
  authenticationMode: API_AND_CONFIG_MAP
apiVersion: eksctl.io/v1alpha5
availabilityZones:
- us-east-1b
- us-east-1d
cloudWatch:
  clusterLogging: {}
fargateProfiles:
- name: fp-default
  selectors:
  - namespace: default
  - namespace: kube-system
  status: ""
iam:
  vpcResourceControllerPolicy: true
  withOIDC: false
kind: ClusterConfig
kubernetesNetworkConfig:
  ipFamily: IPv4
metadata:
  name: webinar-20240314
  region: us-east-1
  version: "1.29"
privateCluster:
  enabled: false
  skipEndpointCreation: false
vpc:
  autoAllocateIPv6: false
  cidr: 192.168.0.0/16
  clusterEndpoints:
    privateAccess: false
    publicAccess: true
  manageSharedNodeSecurityGroupRules: true
  nat:
    gateway: Single

```
