package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/luthermonson/go-proxmox"
)

var (
	_ datasource.DataSource              = &nodesDataSource{}
	_ datasource.DataSourceWithConfigure = &nodesDataSource{}
)

func NewNodesDataSource() datasource.DataSource {
	return &nodesDataSource{}
}

type nodesDataSource struct {
	client *proxmox.Client
}

type nodesDataSourceModel struct {
	Data []nodeModel `tfsdk:"data"`
}

// coffeesModel maps coffees schema data.
type nodeModel struct {
	ID             types.String `tfsdk:"id"`
	Node           types.String `tfsdk:"node"`
	Status         types.String `tfsdk:"status"`
	MaxCPU         types.Int64  `tfsdk:"maxcpu"`
	MaxMem         types.Int64  `tfsdk:"maxmem"`
	MaxDisk        types.Int64  `tfsdk:"maxdisk"`
	SSLFingerprint types.String `tfsdk:"ssl_fingerprint"`
	Type           types.String `tfsdk:"type"`
	Level          types.String `tfsdk:"level"`
}

func (d *nodesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *nodesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nodes"
}

func (d *nodesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"data": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"node": schema.StringAttribute{
							Computed: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
						"maxdisk": schema.Int64Attribute{
							Computed: true,
						},
						"maxcpu": schema.Int64Attribute{
							Computed: true,
						},
						"ssl_fingerprint": schema.StringAttribute{
							Computed: true,
						},
						"maxmem": schema.Int64Attribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"level": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *nodesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state nodesDataSourceModel

	cluster, err := d.client.Cluster(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox Cluster",
			err.Error(),
		)
		return
	}
	nodes := cluster.Nodes

	// Map response body to model
	for _, node := range nodes {
		nodeState := nodeModel{
			ID:             types.StringValue(node.ID),
			Node:           types.StringValue(node.Name),
			Status:         types.StringValue(node.Status),
			MaxCPU:         types.Int64Value(int64(node.MaxCPU)),
			MaxMem:         types.Int64Value(int64(node.MaxMem)),
			MaxDisk:        types.Int64Value(int64(node.MaxDisk)),
			SSLFingerprint: types.StringValue(node.SSLFingerprint),
			Type:           types.StringValue(node.Type),
			Level:          types.StringValue(node.Level),
		}

		state.Data = append(state.Data, nodeState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
