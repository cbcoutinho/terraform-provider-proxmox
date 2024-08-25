package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/luthermonson/go-proxmox"
)

var (
	_ datasource.DataSource              = &nodeNetworksDataSource{}
	_ datasource.DataSourceWithConfigure = &nodeNetworksDataSource{}
)

func NewNodeNetworksDataSource() datasource.DataSource {
	return &nodeNetworksDataSource{}
}

type nodeNetworksDataSource struct {
	client *proxmox.Client
}

type nodeNetworkModel struct {
	Active types.Bool   `tfsdk:"active"`
	Method types.String `tfsdk:"method"`
	Iface  types.String `tfsdk:"iface"`
	Type   types.String `tfsdk:"type"`
}

type nodeNetworksDataSourceModel struct {
	Node     types.String       `tfsdk:"node"`
	Networks []nodeNetworkModel `tfsdk:"networks"`
}

func (d *nodeNetworksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*proxmox.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *proxmox.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *nodeNetworksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_networks"
}

func (d *nodeNetworksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"node": schema.StringAttribute{
				Required: true,
			},
			"networks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"active": schema.BoolAttribute{
							Computed: true,
						},
						"method": schema.StringAttribute{
							Computed: true,
						},
						"iface": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *nodeNetworksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state nodeNetworksDataSourceModel

	// Get `node` from configuration
	var nodeName types.String
	diags := req.Config.GetAttribute(ctx, path.Root("node"), &nodeName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	node, err := d.client.Node(ctx, nodeName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox Node",
			err.Error(),
		)
		return
	}

	networks, err := node.Networks(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox Node Networks",
			err.Error(),
		)
		return
	}

	for _, network := range networks {

		active := network.Active == 1 // Convert from int to bool

		networkState := nodeNetworkModel{
			Active: types.BoolValue(active),
			Method: types.StringValue(network.Method),
			Iface:  types.StringValue(network.Iface),
			Type:   types.StringValue(network.Type),
		}

		state.Networks = append(state.Networks, networkState)
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
