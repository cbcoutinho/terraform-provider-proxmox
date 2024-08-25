package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/luthermonson/go-proxmox"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &proxmoxProvider{}
)

// proxmoxProviderModel maps provider schema data to a Go type.
type proxmoxProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &proxmoxProvider{
			version: version,
		}
	}
}

// proxmoxProvider is the provider implementation.
type proxmoxProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *proxmoxProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "proxmox"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *proxmoxProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "URI for Proxmox VE API. May also be provided via PROXMOX_HOST environment variable.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username for Proxmox VE API. May also be provided via PROXMOX_USERNAME environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for Proxmox VE API. May also be provided via PROXMOX_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a proxmox API client for data sources and resources.
func (p *proxmoxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config proxmoxProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Proxmox VE API Host",
			"The provider cannot create the Proxmox VE API client as there is an unknown configuration value for the Proxmox VE API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PROXMOX_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Proxmox VE API Username",
			"The provider cannot create the Proxmox VE API client as there is an unknown configuration value for the Proxmox VE API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PROXMOX_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Proxmox VE API Password",
			"The provider cannot create the Proxmox VE API client as there is an unknown configuration value for the Proxmox VE API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PROXMOX_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("PROXMOX_HOST")
	username := os.Getenv("PROXMOX_USERNAME")
	password := os.Getenv("PROXMOX_PASSWORD")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Proxmox VE API Host",
			"The provider cannot create the Proxmox VE API client as there is a missing or empty value for the Proxmox VE API host. "+
				"Set the host value in the configuration or use the PROXMOX_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Proxmox VE API Username",
			"The provider cannot create the Proxmox VE API client as there is a missing or empty value for the Proxmox VE API username. "+
				"Set the username value in the configuration or use the PROXMOX_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Proxmox VE API Password",
			"The provider cannot create the Proxmox VE API client as there is a missing or empty value for the Proxmox VE API password. "+
				"Set the password value in the configuration or use the PROXMOX_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "proxmox_host", host)
	ctx = tflog.SetField(ctx, "proxmox_username", username)
	ctx = tflog.SetField(ctx, "proxmox_password", password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "proxmox_password")

	tflog.Debug(ctx, "Creating Proxmox VE client")

	insecureHTTPClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	credentials := proxmox.Credentials{
		Username: username,
		Password: password,
	}
	client := proxmox.NewClient(fmt.Sprintf("%s/api2/json", host),
		proxmox.WithHTTPClient(&insecureHTTPClient),
		//proxmox.WithAPIToken(tokenID, secret),
		proxmox.WithCredentials(&credentials),
		proxmox.WithLogger(&proxmox.LeveledLogger{
			Level: proxmox.LevelDebug,
		}),
	)

	version, err := client.Version(ctx)
	if err != nil {
		panic(err)
	}
	tflog.Info(ctx, fmt.Sprintf("Proxmox VE Version: %s", version.Release))

	// Make the Proxmox VE client and cluster available during DataSource and
	// Resource type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *proxmoxProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewNodesDataSource,
		NewNodeNetworksDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *proxmoxProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClusterFirewallGroupResource,
	}
}
