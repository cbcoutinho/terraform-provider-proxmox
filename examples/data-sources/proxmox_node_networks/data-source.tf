terraform {
  required_providers {
    proxmox = {
      source = "cbcoutinho/proxmox"
    }
  }
}

provider "proxmox" {}

data "proxmox_node_networks" "proxmox" {
  node = "proxmox"
}

output "proxmox_node_networks" {
  value = data.proxmox_node_networks.proxmox
}
