terraform {
  required_providers {
    proxmox = {
      source = "cbcoutinho/proxmox"
    }
  }
}

provider "proxmox" {}

data "proxmox_nodes" "all" {}

output "proxmox_nodes" {
  value = data.proxmox_nodes.all
}
