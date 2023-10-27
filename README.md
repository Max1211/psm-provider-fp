# psm-provider


## Installation
Currently the code is installed directly from Github where there is a Main repo and a Dev repository. Current expections are that the provider is being installed using a Linux or WSL based platform which it has been tested against. You will need to have git tools installed and a functional build system. 

```
git clone https://github.com/farsonic/psm-provider.git 
cd psm-provider
make
```

Once you install the provider it will be hosted locally with the current **Hostname = local** and the **Namespace = provider**. The Name of the provider is **PSM**.

Within your Terraform infrastructure file (ie main.tf) specify the provider with the following syntax. 

```
terraform { 
  required_providers {
   psm = { 
      version = "0.1.81" 
      source = "local/provider/psm"
  }
}
```

You will need to configure the provider to communicate directly with the PSM server either with or without SSL certificate validation using the following definition. 

```
provider "psm" { 
  user = "admin"
  server = "https://PSM_SERVER"
  password = "PSM_PASSWORD"
  insecure = true
}
```

## Usage examples

### Network 
Within PSM a network definition defines the name of the network and the VLAN that will be redirected to a DPU. The following resource definition will create a network called "Database Network" which redirect VLAN 123 traffic to the DPU. 

```
resource "psm_network" "network" {
  name     = "Database Network"
  tenant   = "default" 
  vlan_id  = 123
}
```

### IP Collections
PSM allows the user to create groups of IP Addresses called IP Collections. These are then used within Security Policies (and elsewhere) to define the source and destination IP Addresses used for matches. 

```
resource "psm_ipcollection" "ipcollections" {
  name     = "DatabaseServers"
  addresses = "10.10.10.0/24" 
}
```

### Security Policies 
Security policies are attached to either an individual network or to the VRF. If attached to a VRF then the policy is inherited by networks associated with that particular VRF. If the tenant and/or the policy_distribution_target is not defined these will default to the default VRF. The from IP_Collections will need to be defined prior to them being mapped within the rule. 

```
resource "psm_rules" "default_vrf_policy" {
  policy_name                = "AllowSSH"
  tenant                     = "default"
  policy_distribution_target = "default"
  rule {
      rule_name = "AllowSSHTraffic"
      description = "This rule allows SSH traffic from public IPs"
      from_ip_collections = ["network2","network1"]
      to_ip_collections   = ["network4"]
      apps = ["SSH"]
      action = "permit"
    }
```



