terraform {
  required_providers {
    proxmox = {
      source = "cbcoutinho/proxmox"
    }
  }
}

provider "proxmox" {}

resource "proxmox_sdn_zone" "example" {
  zone = "example"

  #type = "simple"

  type   = "vlan"
  bridge = "vmbr0"

  #dns    = "192.168.2.201"
}

output "sdn_zone" {
  value = proxmox_sdn_zone.example
}
