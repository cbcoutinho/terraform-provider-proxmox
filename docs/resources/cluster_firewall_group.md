---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "proxmox_cluster_firewall_group Resource - proxmox"
subcategory: ""
description: |-
  
---

# proxmox_cluster_firewall_group (Resource)



## Example Usage

```terraform
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
  #comment = "test comment" # TODO: Not yet supported
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `group` (String)

### Optional

- `rules` (Attributes List) (see [below for nested schema](#nestedatt--rules))

<a id="nestedatt--rules"></a>
### Nested Schema for `rules`

Required:

- `action` (String)
- `type` (String)
