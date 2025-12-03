package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/vimaster/terraform-provider-allinkl/internal/allinkl"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &recordsResource{}
	_ resource.ResourceWithConfigure   = &recordsResource{}
	_ resource.ResourceWithImportState = &recordsResource{}
)

// NewRecordsResource is a helper function to simplify the provider implementation.
func NewRecordsResource() resource.Resource {
	return &recordsResource{}
}

// recordsResource is the resource implementation.
type recordsResource struct {
	client *allinkl.Client
}

// Metadata returns the resource type name.
func (r *recordsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_records"
}

// recordsResourceModel maps the resource schema data.
type recordsResourceModel struct {
	RecordId    types.Int64  `tfsdk:"record_id"` // synthetic key like dyndns_login
	LastUpdated types.String `tfsdk:"last_updated"`
	ZoneHost    types.String `tfsdk:"zone_host"`
	RecordType  types.String `tfsdk:"record_type"`
	RecordName  types.String `tfsdk:"record_name"`
	RecordData  types.String `tfsdk:"record_data"`
	RecordAux   types.Int64  `tfsdk:"record_aux"`
}

// Schema defines the schema for the resource.
func (r *recordsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"record_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

func (d *recordsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func ValidateRecord(recordType string, recordAux int64) bool {
	if recordType != "MX" && recordType != "mx" && recordType != "SRV" && recordType != "srv" {
		if recordAux != 0 {
			return false
		}
	}
	return true
}

// Create creates the resource and sets the initial Terraform state.
func (r *recordsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan recordsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate aux: only allowed non-zero for MX or SRV. Otherwise must be 0.
	if !ValidateRecord(plan.RecordType.ValueString(), plan.RecordAux.ValueInt64()) {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationCreate, "record_aux must be empty or 0 unless record_type is MX or SRV")
		return
	}

	recReq := allinkl.RecordRequest{
		ZoneHost:   plan.ZoneHost.ValueString(),
		RecordType: plan.RecordType.ValueString(),
		RecordName: plan.RecordName.ValueString(),
		RecordData: plan.RecordData.ValueString(),
		RecordAux:  plan.RecordAux.ValueInt64(),
	}

	recordId, err := r.client.AddRecord(ctx, recReq)
	if err != nil {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationCreate, err.Error())
		return
	}

	plan.RecordId = types.Int64Value(recordId)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	plan.ZoneHost = types.StringValue(plan.ZoneHost.ValueString())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *recordsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state recordsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	records, err := r.client.GetRecords(ctx, state.ZoneHost.ValueString())
	if err != nil {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationRead, err.Error())
		return
	}

	var record allinkl.GetRecordReturnInfo
	for _, r := range records {
		recordAsInt64, err := strconv.ParseInt(string(r.RecordId), 10, 64)
		if err != nil {
			continue
		}
		if recordAsInt64 == state.RecordId.ValueInt64() {
			record = allinkl.GetRecordReturnInfo{
				RecordZone: r.RecordZone,
				RecordType: r.RecordType,
				RecordName: r.RecordName,
				RecordData: r.RecordData,
				RecordAux:  r.RecordAux,
			}
			break
		}
	}

	if record.RecordName == "" {
		resp.Diagnostics.AddError(
			"Record Not Found",
			fmt.Sprintf("No record found for login: %d", state.RecordId.ValueInt64()),
		)
		return
	}

	state = recordsResourceModel{
		RecordId:    types.Int64Value(state.RecordId.ValueInt64()),
		LastUpdated: state.LastUpdated,
		ZoneHost:    types.StringValue(record.RecordZone),
		RecordType:  types.StringValue(record.RecordType),
		RecordName:  types.StringValue(record.RecordName),
		RecordData:  types.StringValue(record.RecordData),
		RecordAux:   types.Int64Value(record.RecordAux),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *recordsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan recordsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	recUpd := allinkl.RecordUpdateRequest{
		RecordId:   plan.RecordId.ValueInt64(),
		ZoneHost:   plan.ZoneHost.ValueString(),
		RecordType: plan.RecordType.ValueString(),
		RecordName: plan.RecordName.ValueString(),
		RecordData: plan.RecordData.ValueString(),
		RecordAux:  plan.RecordAux.ValueInt64(),
	}

	// Validate aux: only allowed non-zero for MX or SRV. Otherwise must be 0.
	if !ValidateRecord(plan.RecordType.ValueString(), plan.RecordAux.ValueInt64()) {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationCreate, "record_aux must be empty or 0 unless record_type is MX or SRV")
		return
	}

	status, err := r.client.UpdateRecord(ctx, recUpd)
	if err != nil {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationUpdate, err.Error())
		return
	}
	if status != "TRUE" {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationUpdate, "UpdateRecord returned unexpected status: "+status)
		return
	}

	records, err := r.client.GetRecords(ctx, plan.ZoneHost.ValueString())
	if err != nil {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationRead, err.Error())
		return
	}
	var record allinkl.GetRecordReturnInfo
	for _, r := range records {
		recordAsInt64, err := strconv.ParseInt(string(r.RecordId), 10, 64)
		if err != nil {
			continue
		}
		if recordAsInt64 == plan.RecordId.ValueInt64() {
			record = r
			break
		}
	}

	recordAsInt64, err := strconv.ParseInt(string(record.RecordId), 10, 64)
	if err != nil {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationRead, "Could not parse record ID: "+err.Error())
		return
	}
	plan = recordsResourceModel{
		RecordId:    types.Int64Value(recordAsInt64),
		LastUpdated: types.StringValue(time.Now().Format(time.RFC850)),
		ZoneHost:    types.StringValue(record.RecordZone),
		RecordType:  types.StringValue(record.RecordType),
		RecordName:  types.StringValue(record.RecordName),
		RecordData:  types.StringValue(record.RecordData),
		RecordAux:   types.Int64Value(record.RecordAux),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *recordsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state recordsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleted, err := r.client.DeleteRecord(ctx, state.RecordId.ValueInt64())
	if deleted != "" {
		resp.Diagnostics.AddError(
			"Error Deleting AllInkl Records",
			fmt.Sprintf("Could not delete record; received responseInfo of '%s': %v", deleted, err),
		)
		return
	}
}

func (r *recordsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID to be the record_id.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
