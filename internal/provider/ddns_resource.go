package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/vimaster/terraform-provider-allinkl/internal/allinkl"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ddnsResource{}
	_ resource.ResourceWithConfigure   = &ddnsResource{}
	_ resource.ResourceWithImportState = &ddnsResource{}
)

// NewDDNSResource is a helper function to simplify the provider implementation.
func NewDDNSResource() resource.Resource {
	return &ddnsResource{}
}

// ddnsResource is the resource implementation.
type ddnsResource struct {
	client *allinkl.Client
}

// Metadata returns the resource type name.
func (r *ddnsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ddns"
}

// ddnsResourceModel maps the resource schema data.
type ddnsResourceModel struct {
	DyndnsLogin    types.String `tfsdk:"dyndns_login"` // PRIMARY KEY
	LastUpdated    types.String `tfsdk:"last_updated"`
	DyndnsComment  types.String `tfsdk:"dyndns_comment"`
	DyndnsPassword types.String `tfsdk:"dyndns_password"`
	DyndnsZone     types.String `tfsdk:"dyndns_zone"`
	DyndnsLabel    types.String `tfsdk:"dyndns_label"`
	DyndnsTargetIP types.String `tfsdk:"dyndns_target_ip"`
}

// Schema defines the schema for the resource.
func (r *ddnsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"dyndns_login": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"dyndns_comment": schema.StringAttribute{
				Required: true,
			},
			"dyndns_password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"dyndns_zone": schema.StringAttribute{
				Required: true,
			},
			"dyndns_label": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dyndns_target_ip": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *ddnsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*allinkl.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *allinkl.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *ddnsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ddnsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ddnsReq := allinkl.DDNSRequest{
		DyndnsComment:  plan.DyndnsComment.ValueString(),
		DyndnsPassword: plan.DyndnsPassword.ValueString(),
		DyndnsZone:     plan.DyndnsZone.ValueString(),
		DyndnsLabel:    plan.DyndnsLabel.ValueString(),
		DyndnsTargetIP: plan.DyndnsTargetIP.ValueString(),
	}

	login, err := r.client.AddDDNSUser(ctx, ddnsReq)
	if err != nil {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationCreate, err.Error())
		return
	}

	plan.DyndnsLogin = types.StringValue(login) // dyndns_login
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ddnsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ddnsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ddns, err := r.client.GetDDNSUser(ctx, state.DyndnsLogin.ValueString())
	if err != nil {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationRead, err.Error())
		return
	}

	state = ddnsResourceModel{
		DyndnsLogin:    types.StringValue(ddns.DyndnsLogin),
		LastUpdated:    state.LastUpdated,
		DyndnsComment:  types.StringValue(ddns.DyndnsComment),
		DyndnsPassword: types.StringValue(ddns.DyndnsPassword),
		DyndnsZone:     types.StringValue(ddns.DyndnsZone),
		DyndnsLabel:    types.StringValue(ddns.DyndnsLabel),
		DyndnsTargetIP: types.StringValue(ddns.DyndnsTargetIP),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ddnsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ddnsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ddnsReq := allinkl.DDNSUpdateRequest{
		DyndnsLogin:    plan.DyndnsLogin.ValueString(),
		DyndnsComment:  plan.DyndnsComment.ValueString(),
		DyndnsPassword: plan.DyndnsPassword.ValueString(),
		DyndnsZone:     plan.DyndnsZone.ValueString(),
		DyndnsLabel:    plan.DyndnsLabel.ValueString(),
		DyndnsTargetIP: plan.DyndnsTargetIP.ValueString(),
	}

	status, err := r.client.UpdateDDNSUser(ctx, ddnsReq)
	if err != nil {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationUpdate, err.Error())
		return
	}
	if status != "TRUE" {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationUpdate, "UpdateDDNSUser returned unexpected status: "+status)
		return
	}

	ddns, err := r.client.GetDDNSUser(ctx, plan.DyndnsLogin.ValueString())
	if err != nil {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationRead, err.Error())
		return
	}

	plan = ddnsResourceModel{
		DyndnsLogin:    types.StringValue(ddns.DyndnsLogin),
		LastUpdated:    types.StringValue(time.Now().Format(time.RFC850)),
		DyndnsComment:  types.StringValue(ddns.DyndnsComment),
		DyndnsPassword: types.StringValue(ddns.DyndnsPassword),
		DyndnsZone:     types.StringValue(ddns.DyndnsZone),
		DyndnsLabel:    types.StringValue(ddns.DyndnsLabel),
		DyndnsTargetIP: types.StringValue(ddns.DyndnsTargetIP),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ddnsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ddnsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleted, err := r.client.DeleteDDNSUser(ctx, state.DyndnsLogin.ValueString())
	if deleted != "" {
		resp.Diagnostics.AddError(
			"Error Deleting AllInkl DDNS",
			fmt.Sprintf("Could not delete ddns; received responseInfo of '%s': %v", deleted, err),
		)
		return
	}
}

func (r *ddnsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by dyndns_login
	if req.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID to be the dyndns_login.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dyndns_login"), req.ID)...)
}
