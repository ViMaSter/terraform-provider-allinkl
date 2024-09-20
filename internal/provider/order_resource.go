package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-allinkl/internal/allinkl"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &dnsResource{}
	_ resource.ResourceWithConfigure   = &dnsResource{}
	_ resource.ResourceWithImportState = &dnsResource{}
)

// NewDNSResource is a helper function to simplify the provider implementation.
func NewDNSResource() resource.Resource {
	return &dnsResource{}
}

// dnsResource is the resource implementation.
type dnsResource struct {
	client *allinkl.Client
}

// Metadata returns the resource type name.
func (r *dnsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns"
}

// dnsResourceModel maps the resource schema data.
type dnsResourceModel struct {
	ID          types.String `tfsdk:"id"`
	LastUpdated types.String `tfsdk:"last_updated"`
	ZoneHost    types.String `tfsdk:"zone_host"`
	RecordType  types.String `tfsdk:"record_type"`
	RecordName  types.String `tfsdk:"record_name"`
	RecordData  types.String `tfsdk:"record_data"`
	RecordAux   types.Int64  `tfsdk:"record_aux"`
}

// Schema defines the schema for the resource.
func (r *dnsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"zone_host": schema.StringAttribute{
				Required: true,
			},
			"record_type": schema.StringAttribute{
				Required: true,
			},
			"record_name": schema.StringAttribute{
				Required: true,
			},
			"record_data": schema.StringAttribute{
				Required: true,
			},
			"record_aux": schema.Int64Attribute{
				Required: true,
			},
		},
	}
}

func (d *dnsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
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
// Create a new resource.
func (r *dnsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan dnsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var allinklItem = allinkl.DNSRequest{
		ZoneHost:   plan.ZoneHost.ValueString(),
		RecordType: plan.RecordType.ValueString(),
		RecordName: plan.RecordName.ValueString(),
		RecordData: plan.RecordData.ValueString(),
		RecordAux:  int(plan.RecordAux.ValueInt64()),
	}

	id, err := r.client.AddDNSSettings(ctx, allinklItem)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating AllInkl DNS",
			"Could not create dns, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(id)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
// Read resource information.
func (r *dnsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state dnsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed dns value from AllInkl
	dns, err := r.client.GetDNSSettings(ctx, state.ZoneHost.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AllInkl DNS",
			"Could not read AllInkl dns ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	var dnsCount int = len(dns)
	if dnsCount == 0 {
		resp.Diagnostics.AddError(
			"Error Reading AllInkl DNS",
			"Could not read AllInkl dns ID "+state.ID.ValueString()+": no records found, expected 1",
		)
		return
	}

	if dnsCount > 1 {
		resp.Diagnostics.AddError(
			"Error Reading AllInkl DNS",
			fmt.Sprintf("Could not read AllInkl dns ID %s: found %d records, expected 1", state.ID.ValueString(), dnsCount),
		)
		return
	}

	state = dnsResourceModel{
		ID:         state.ID,
		ZoneHost:   types.StringValue(dns[0].ZoneHost),
		RecordType: types.StringValue(dns[0].RecordType),
		RecordName: types.StringValue(dns[0].RecordName),
		RecordData: types.StringValue(dns[0].RecordData),
		RecordAux:  types.Int64Value(int64(dns[0].RecordAux)),
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *dnsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan dnsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var allinklItem = allinkl.DNSRequest{
		RecordId:   plan.ID.ValueString(),
		ZoneHost:   plan.ZoneHost.ValueString(),
		RecordType: plan.RecordType.ValueString(),
		RecordName: plan.RecordName.ValueString(),
		RecordData: plan.RecordData.ValueString(),
		RecordAux:  int(plan.RecordAux.ValueInt64()),
	}

	_, err := r.client.UpdateDNSSettings(ctx, allinklItem)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating AllInkl DNS",
			"Could not update dns, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	dns, err := r.client.GetDNSSettings(ctx, plan.ZoneHost.ValueString(), plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AllInkl DNS",
			"Could not read AllInkl dns ID "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	var dnsCount int = len(dns)
	if dnsCount == 0 {
		resp.Diagnostics.AddError(
			"Error Reading AllInkl DNS",
			"Could not read AllInkl dns ID "+plan.ID.ValueString()+": no records found, expected 1",
		)
		return
	}

	if dnsCount > 1 {
		resp.Diagnostics.AddError(
			"Error Reading AllInkl DNS",
			fmt.Sprintf("Could not read AllInkl dns ID %s: found %d records, expected 1", plan.ID.ValueString(), dnsCount),
		)
		return
	}

	plan = dnsResourceModel{
		ID:          plan.ID,
		LastUpdated: types.StringValue(time.Now().Format(time.RFC850)),
		ZoneHost:    types.StringValue(dns[0].ZoneHost),
		RecordType:  types.StringValue(dns[0].RecordType),
		RecordName:  types.StringValue(dns[0].RecordName),
		RecordData:  types.StringValue(dns[0].RecordData),
		RecordAux:   types.Int64Value(int64(dns[0].RecordAux)),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *dnsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state dnsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleted, err := r.client.DeleteDNSSettings(ctx, state.ID.ValueString())
	if !deleted {
		resp.Diagnostics.AddError(
			"Error Deleting AllInkl DNS",
			"Could not delete dns, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *dnsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if path.Root("id").Equal(path.Empty()) {
		resp.Diagnostics.AddError(
			"Resource Import Passthrough Missing Attribute Path",
			"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				"Resource ImportState method call to ImportStatePassthroughID path must be set to a valid attribute path that can accept a string value.",
		)
	}

	// split into zone_host and record_id by `/`
	var zoneHost, recordID string
	if req.ID != "" {
		zoneHost, recordID = req.ID, ""
		if i := strings.Index(req.ID, "/"); i != -1 {
			zoneHost, recordID = req.ID[:i], req.ID[i+1:]
		}
	}

	if recordID == "" || zoneHost == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in the format `zone_host/record_id`, got: "+req.ID,
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone_host"), zoneHost)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), recordID)...)
}
