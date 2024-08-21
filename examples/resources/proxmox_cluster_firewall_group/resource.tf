terraform {
  required_providers {
    proxmox = {
      source = "cbcoutinho/proxmox"
    }
  }
}

provider "proxmox" {}

resource "proxmox_cluster_firewall_group" "example" {
  group = "example"
}
