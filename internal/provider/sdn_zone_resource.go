package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/luthermonson/go-proxmox"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &sdnZoneResource{}
	_ resource.ResourceWithConfigure = &sdnZoneResource{}
)

// NewClusterFirewallGroupResource is a helper function to simplify the provider implementation.
func NewSdnZoneResource() resource.Resource {
	return &sdnZoneResource{}
}

// clusterFirewallGroupResource is the resource implementation.
type sdnZoneResource struct {
	client *proxmox.Client
}

type sdnZoneResourceModel struct {
	Zone   types.String `tfsdk:"zone"`
	Type   types.String `tfsdk:"type"`
	Dns    types.String `tfsdk:"dns"`
	Bridge types.String `tfsdk:"bridge"`
	Digest types.String `tfsdk:"digest"`
}

func (z *sdnZoneResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	z.client = client
}

// Metadata returns the resource type name.
func (z *sdnZoneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sdn_zone"
}

func (z *sdnZoneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"zone": schema.StringAttribute{
				Required:    true,
				Description: "The SDN zone object identifier",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Plugin type",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("evpn", "faucet", "qinq", "simple", "vlan", "vxlan"),
				},
			},
			"dns": schema.StringAttribute{
				Optional: true,
			},
			"bridge": schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{}, // TODO: Make `bridge` required if `type` == "vlan"
			},
			"digest": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (z *sdnZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	tflog.Info(ctx, "Getting data from plan for proxmox_sdn_zone")
	var plan sdnZoneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Mapping schema resource attributes to API params")
	zoneName := plan.Zone.ValueString()
	config := proxmox.ZoneConfig{
		Zone:   plan.Zone.ValueString(),
		Type:   plan.Type.ValueString(),
		Dns:    plan.Dns.ValueString(),
		Bridge: plan.Bridge.ValueString(),
	}

	tflog.Info(ctx, fmt.Sprintf("Creating new zone %s", zoneName))
	zone, err := z.client.NewZone(ctx, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox Node",
			err.Error(),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Mapping response values to resource schema %s", zone.Zone))
	plan.Zone = types.StringValue(zone.Zone)
	plan.Type = types.StringValue(zone.Type)
	plan.Digest = types.StringValue(zone.Digest)

	if zone.Dns != "" {
		plan.Dns = types.StringValue(zone.Dns)
	}
	if zone.Bridge != "" {
		plan.Bridge = types.StringValue(zone.Bridge)
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully created Zone: %s", zone.Zone))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (z *sdnZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	tflog.Info(ctx, "Reading SDN Zone state")
	var state sdnZoneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Refreshing SDN Zone from Proxmox")
	zoneName := state.Zone.ValueString()
	ret, err := z.client.Zone(ctx, zoneName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox Node",
			err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Mapping response values to resource schema attributes")
	state.Zone = types.StringValue(ret.Zone)
	state.Type = types.StringValue(ret.Type)
	state.Bridge = types.StringValue(ret.Bridge)
	if ret.Dns != "" {
		state.Dns = types.StringValue(ret.Dns)
	}
	if ret.Bridge != "" {
		state.Bridge = types.StringValue(ret.Bridge)
	}

	state.Dns = types.StringValue(ret.Dns)
	state.Bridge = types.StringValue(ret.Bridge)

	tflog.Info(ctx, "Setting state based on response from Proxmox")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (z *sdnZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Reading SDN Zone config from plan")
	var plan sdnZoneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Generating API request from plan")
	zoneName := plan.Zone.ValueString()
	config := proxmox.ZoneConfig{
		Zone:   plan.Zone.ValueString(),
		Dns:    plan.Dns.ValueString(),
		Bridge: plan.Bridge.ValueString(),
	}
	ret, err := z.client.Zone(ctx, zoneName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox Node",
			err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Updating the SDN Zone")
	ret.Update(ctx, config)

	tflog.Info(ctx, "Mapping response to resource schema attributes")
	ret, err = z.client.Zone(ctx, zoneName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox Node",
			err.Error(),
		)
		return
	}
	plan.Zone = types.StringValue(ret.Zone)
	plan.Type = types.StringValue(ret.Type)
	plan.Digest = types.StringValue(ret.Digest)
	if ret.Dns != "" {
		plan.Dns = types.StringValue(ret.Dns)
	}
	if ret.Bridge != "" {
		plan.Bridge = types.StringValue(ret.Bridge)
	}

	tflog.Info(ctx, "Set state with the updated SDN Zone")

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (z *sdnZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Reading SDN Zone config from state")
	var state sdnZoneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting SDN Zone")
	zone, err := z.client.Zone(ctx, state.Zone.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox Node",
			err.Error(),
		)
		return
	}

	zone.Delete(ctx)

	return
}
