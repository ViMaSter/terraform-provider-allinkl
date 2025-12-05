package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/vimaster/terraform-provider-allinkl/internal/allinkl"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	RecordId    types.Int64  `tfsdk:"record_id"` // synthetic key like dyndns_login
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
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "Priority value for MX and SRV records. Optional for all other record types (defaults to 0). For non-MX/SRV records, this value must be 0 or omitted; non-zero values will result in a validation error.",
			},
		},
	}
}

func (r *dnsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*allinkl.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *allinkl.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func ValidateDNSRecord(recordType string, recordAux int64) bool {
	// Normalize record type to uppercase for consistent comparison
	normalizedType := strings.ToUpper(recordType)
	if normalizedType != "MX" && normalizedType != "SRV" {
		if recordAux != 0 {
			return false
		}
	}
	return true
}

// Create creates the resource and sets the initial Terraform state.
func (r *dnsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate aux: only allowed non-zero for MX or SRV. Otherwise must be 0.
	if !ValidateDNSRecord(plan.RecordType.ValueString(), plan.RecordAux.ValueInt64()) {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationCreate, "record_aux must be empty or 0 unless record_type is MX or SRV")
		return
	}

	recReq := allinkl.DNSRequest{
		ZoneHost:   plan.ZoneHost.ValueString(),
		RecordType: plan.RecordType.ValueString(),
		RecordName: plan.RecordName.ValueString(),
		RecordData: plan.RecordData.ValueString(),
		RecordAux:  plan.RecordAux.ValueInt64(),
	}

	recordId, err := r.client.AddDNS(ctx, recReq)
	if err != nil {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationCreate, err.Error())
		return
	}

	plan.RecordId = types.Int64Value(recordId)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *dnsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dnsRecords, err := r.client.GetDNS(ctx, state.ZoneHost.ValueString())
	if err != nil {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationRead, err.Error())
		return
	}

	var dns allinkl.GetDNSReturnInfo
	for _, r := range dnsRecords {
		recordAsInt64, err := strconv.ParseInt(string(r.RecordId), 10, 64)
		if err != nil {
			continue
		}
		if recordAsInt64 == state.RecordId.ValueInt64() {
			dns = allinkl.GetDNSReturnInfo{
				RecordZone: r.RecordZone,
				RecordType: r.RecordType,
				RecordName: r.RecordName,
				RecordData: r.RecordData,
				RecordAux:  r.RecordAux,
			}
			break
		}
	}

	if dns.RecordName == "" {
		resp.Diagnostics.AddError(
			"DNS Not Found",
			fmt.Sprintf("No record found for record_id: %d", state.RecordId.ValueInt64()),
		)
		return
	}

	state = dnsResourceModel{
		RecordId:    types.Int64Value(state.RecordId.ValueInt64()),
		LastUpdated: state.LastUpdated,
		ZoneHost:    types.StringValue(dns.RecordZone),
		RecordType:  types.StringValue(dns.RecordType),
		RecordName:  types.StringValue(dns.RecordName),
		RecordData:  types.StringValue(dns.RecordData),
		RecordAux:   types.Int64Value(dns.RecordAux),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *dnsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dnsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	recUpd := allinkl.DNSUpdateRequest{
		RecordId:   plan.RecordId.ValueInt64(),
		ZoneHost:   plan.ZoneHost.ValueString(),
		RecordType: plan.RecordType.ValueString(),
		RecordName: plan.RecordName.ValueString(),
		RecordData: plan.RecordData.ValueString(),
		RecordAux:  plan.RecordAux.ValueInt64(),
	}

	// Validate aux: only allowed non-zero for MX or SRV. Otherwise must be 0.
	if !ValidateDNSRecord(plan.RecordType.ValueString(), plan.RecordAux.ValueInt64()) {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationUpdate, "record_aux must be empty or 0 unless record_type is MX or SRV")
		return
	}

	status, err := r.client.UpdateDNS(ctx, recUpd)
	if err != nil {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationUpdate, err.Error())
		return
	}
	if status != "TRUE" {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationUpdate, "UpdateRecord returned unexpected status: "+status)
		return
	}

	dnsRecords, err := r.client.GetDNS(ctx, plan.ZoneHost.ValueString())
	if err != nil {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationRead, err.Error())
		return
	}
	var dns *allinkl.GetDNSReturnInfo
	for _, r := range dnsRecords {
		idAsInt64, err := strconv.ParseInt(string(r.RecordId), 10, 64)
		if err != nil {
			continue
		}
		if idAsInt64 == plan.RecordId.ValueInt64() {
			dns = &r
			break
		}
	}

	if dns == nil {
		resp.Diagnostics.AddError(
			"DNS Not Found",
			fmt.Sprintf("No record found for record_id: %d after update", plan.RecordId.ValueInt64()),
		)
		return
	}

	idAsInt64, err := strconv.ParseInt(string(dns.RecordId), 10, 64)
	if err != nil {
		reportErrorWithDescription(resp.Diagnostics.AddError, OperationRead, "Could not parse record ID: "+err.Error())
		return
	}
	plan = dnsResourceModel{
		RecordId:    types.Int64Value(idAsInt64),
		LastUpdated: types.StringValue(time.Now().Format(time.RFC850)),
		ZoneHost:    types.StringValue(dns.RecordZone),
		RecordType:  types.StringValue(dns.RecordType),
		RecordName:  types.StringValue(dns.RecordName),
		RecordData:  types.StringValue(dns.RecordData),
		RecordAux:   types.Int64Value(dns.RecordAux),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *dnsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleted, err := r.client.DeleteDNS(ctx, state.RecordId.ValueInt64())
	if deleted != "" {
		resp.Diagnostics.AddError(
			"Error Deleting AllInkl Records",
			fmt.Sprintf("Could not delete record; received responseInfo of '%s': %v", deleted, err),
		)
		return
	}
}

func (r *dnsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID to be the record_id.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("record_id"), req.ID)...)
}
