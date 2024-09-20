package provider

import (
	"context"
	"os"
	"terraform-provider-allinkl/internal/allinkl"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &allinklProvider{}
)

// allinklProviderModel maps provider schema data to a Go type.
type allinklProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &allinklProvider{
			version: version,
		}
	}
}

// allinklProvider is the provider implementation.
type allinklProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *allinklProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "allinkl"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
// Schema defines the provider-level schema for configuration data.
func (p *allinklProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares a AllInkl API client for data sources and resources.
func (p *allinklProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring AllInkl client")

	// Retrieve provider data from configuration
	var config allinklProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown AllInkl API Username",
			"The provider cannot create the AllInkl API client as there is an unknown configuration value for the AllInkl API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ALLINKL_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown AllInkl API Password",
			"The provider cannot create the AllInkl API client as there is an unknown configuration value for the AllInkl API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ALLINKL_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	username := os.Getenv("ALLINKL_USERNAME")
	password := os.Getenv("ALLINKL_PASSWORD")

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing AllInkl API Username",
			"The provider cannot create the AllInkl API client as there is a missing or empty value for the AllInkl API username. "+
				"Set the username value in the configuration or use the ALLINKL_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing AllInkl API Password",
			"The provider cannot create the AllInkl API client as there is a missing or empty value for the AllInkl API password. "+
				"Set the password value in the configuration or use the ALLINKL_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "allinkl_username", username)
	ctx = tflog.SetField(ctx, "allinkl_password", password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "allinkl_password")

	tflog.Debug(ctx, "Creating AllInkl client")

	var client = allinkl.NewClient(username, password)

	// Make the AllInkl client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured AllInkl client", map[string]any{"success": true})
}

func (p *allinklProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// NewCoffeesDataSource,
	}
}

func (p *allinklProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDNSResource,
	}
}
