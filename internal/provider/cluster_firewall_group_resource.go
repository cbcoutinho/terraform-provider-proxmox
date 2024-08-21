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
	_ resource.Resource              = &clusterFirewallGroupResource{}
	_ resource.ResourceWithConfigure = &clusterFirewallGroupResource{}
)

// NewClusterFirewallGroupResource is a helper function to simplify the provider implementation.
func NewClusterFirewallGroupResource() resource.Resource {
	return &clusterFirewallGroupResource{}
}

// clusterFirewallGroupResource is the resource implementation.
type clusterFirewallGroupResource struct {
	client *proxmox.Client
}

// clusterFirewallGroupResourceModel maps the resource schema data.
type clusterFirewallGroupResourceModel struct {
	Group         types.String                    `tfsdk:"group"`
	Comment       types.String                    `tfsdk:"comment"`
	FirewallRules []clusterFirewallGroupRuleModel `tfsdk:"rules"`
}

type clusterFirewallGroupRuleModel struct {
	Action types.String `tfsdk:"action"`
	Type   types.String `tfsdk:"type"`
}

// Configure adds the provider configured client to the resource.
func (r *clusterFirewallGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *clusterFirewallGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_firewall_group"
}

// Schema defines the schema for the resource.
func (r *clusterFirewallGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"group": schema.StringAttribute{
				Required: true,
			},
			"comment": schema.StringAttribute{
				Optional: true,
				Computed: true, // Allow switch from `nil` to `""`
			},
			"rules": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Required: true,
						},
						"type": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *clusterFirewallGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan clusterFirewallGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.Cluster(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox Cluster",
			err.Error(),
		)
		return
	}

	fwGroupN := proxmox.FirewallSecurityGroup{
		Group: plan.Group.ValueString(),
	}

	if !plan.Comment.IsNull() {
		fwGroupN.Comment = plan.Comment.ValueString()
	}

	err = cluster.NewFWGroup(ctx, &fwGroupN)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Proxmox Cluster Firewall Group",
			err.Error(),
		)
		return
	}

	fwGroup, err := cluster.FWGroup(ctx, plan.Group.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to retrieve Proxmox Cluster Firewall Group",
			err.Error(),
		)
		return
	}

	plan.Group = types.StringValue(fwGroup.Group)
	plan.Comment = types.StringPointerValue(&fwGroup.Comment)
	for index, rule := range fwGroup.Rules {
		plan.FirewallRules[index] = clusterFirewallGroupRuleModel{
			Action: types.StringValue(rule.Action),
			Type:   types.StringValue(rule.Type),
		}
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *clusterFirewallGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	// Get current state
	var state clusterFirewallGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.Cluster(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox Cluster",
			err.Error(),
		)
		return
	}

	fwGroup, err := cluster.FWGroup(ctx, state.Group.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to retrieve Proxmox Cluster Firewall Group",
			err.Error(),
		)
		return
	}

	state.Group = types.StringValue(fwGroup.Group)
	state.Comment = types.StringPointerValue(&fwGroup.Comment)
	state.FirewallRules = []clusterFirewallGroupRuleModel{} // NOTE: Set value to empty list before populating
	for index, rule := range fwGroup.Rules {
		state.FirewallRules[index] = clusterFirewallGroupRuleModel{
			Action: types.StringValue(rule.Action),
			Type:   types.StringValue(rule.Type),
		}
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *clusterFirewallGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *clusterFirewallGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
