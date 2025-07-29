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
	KasLogin    types.String `tfsdk:"kas_login"`
	KasAuthType types.String `tfsdk:"kas_auth_type"`
	KasAuthData types.String `tfsdk:"kas_auth_data"`
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
func (p *allinklProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"kas_login": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"kas_auth_type": schema.StringAttribute{
				Optional: true,
			},
			"kas_auth_data": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares a AllInkl API client for data sources and resources.
func (p *allinklProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring AllInkl client")

	var config allinklProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.KasLogin.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("kas_login"),
			"Unknown AllInkl KAS Login",
			"The provider cannot create the AllInkl API client as there is an unknown configuration value for the KAS login. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ALLINKL_KAS_LOGIN environment variable.",
		)
	}
	if config.KasAuthType.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("kas_auth_type"),
			"Unknown AllInkl KAS Auth Type",
			"The provider cannot create the AllInkl API client as there is an unknown configuration value for the KAS auth type. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ALLINKL_KAS_AUTH_TYPE environment variable.",
		)
	}
	if config.KasAuthData.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("kas_auth_data"),
			"Unknown AllInkl KAS Auth Data",
			"The provider cannot create the AllInkl API client as there is an unknown configuration value for the KAS auth data. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ALLINKL_KAS_AUTH_DATA environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	kasLogin := os.Getenv("ALLINKL_KAS_LOGIN")
	kasAuthType := os.Getenv("ALLINKL_KAS_AUTH_TYPE")
	kasAuthData := os.Getenv("ALLINKL_KAS_AUTH_DATA")

	if !config.KasLogin.IsNull() {
		kasLogin = config.KasLogin.ValueString()
	}
	if !config.KasAuthType.IsNull() {
		kasAuthType = config.KasAuthType.ValueString()
	}
	if !config.KasAuthData.IsNull() {
		kasAuthData = config.KasAuthData.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if kasLogin == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("kas_login"),
			"Missing AllInkl KAS Login",
			"The provider cannot create the AllInkl API client as there is a missing or empty value for the KAS login. "+
				"Set the kas_login value in the configuration or use the ALLINKL_KAS_LOGIN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if kasAuthType == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("kas_auth_type"),
			"Missing AllInkl KAS Auth Type",
			"The provider cannot create the AllInkl API client as there is a missing or empty value for the KAS auth type. "+
				"Set the kas_auth_type value in the configuration or use the ALLINKL_KAS_AUTH_TYPE environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if kasAuthData == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("kas_auth_data"),
			"Missing AllInkl KAS Auth Data",
			"The provider cannot create the AllInkl API client as there is a missing or empty value for the KAS auth data. "+
				"Set the kas_auth_data value in the configuration or use the ALLINKL_KAS_AUTH_DATA environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "allinkl_kas_login", kasLogin)
	ctx = tflog.SetField(ctx, "allinkl_kas_auth_type", kasAuthType)
	ctx = tflog.SetField(ctx, "allinkl_kas_auth_data", kasAuthData)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "allinkl_kas_auth_data")

	tflog.Debug(ctx, "Creating AllInkl client")

	var client = allinkl.NewClient(kasLogin, kasAuthType, kasAuthData)

	// Make the AllInkl client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured AllInkl client", map[string]any{"success": true})
}

func (p *allinklProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// DDNSResource,
	}
}

func (p *allinklProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDDNSResource,
	}
}
