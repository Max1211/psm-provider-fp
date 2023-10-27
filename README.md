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

