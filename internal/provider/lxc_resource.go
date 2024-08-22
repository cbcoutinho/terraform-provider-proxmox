package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/luthermonson/go-proxmox"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &lxcResource{}
	_ resource.ResourceWithConfigure = &lxcResource{}
)

// NewLxcResource is a helper function to simplify the provider implementation.
func NewLxcResource() resource.Resource {
	return &lxcResource{}
}

// lxcResource is the resource implementation.
type lxcResource struct {
	client *proxmox.Client
}

// lxcResourceModel maps the resource schema data.
type lxcResourceModel struct {
	Node       types.String `tfsdk:"node"`
	OSTemplate types.String `tfsdk:"os_template"`
	VMID       types.Int64  `tfsdk:"vm_id"`
}

// Configure adds the provider configured client to the resource.
func (r *lxcResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

// Metadata returns the resource type name.
func (r *lxcResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lxc"
}

// Schema defines the schema for the resource.
func (r *lxcResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"node": schema.StringAttribute{
				Required: true,
			},
			"os_template": schema.StringAttribute{
				Required: true,
			},
			"vm_id": schema.Int64Attribute{
				Required: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *lxcResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan lxcResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Cluster(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox Cluster",
			err.Error(),
		)
		return
	}

	/*
		TODO: There is currently no API for creating LXC clusters
	*/

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *lxcResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *lxcResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *lxcResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
