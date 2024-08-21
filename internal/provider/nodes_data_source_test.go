package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNodesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccNodesDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.proxmox_nodes.all", "id", "example-id"),
				),
			},
		},
	})
}

const testAccNodesDataSourceConfig = `
data "proxmox_nodes" "all" {
  configurable_attribute = "example"
}
`
