package provider

import (
	"context"
	"errors"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &MonitorResource{}
	_ resource.ResourceWithConfigure   = &MonitorResource{}
	_ resource.ResourceWithImportState = &MonitorResource{}
)

type MonitorResource struct {
	client *SporkClient
}

type MonitorResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Target          types.String `tfsdk:"target"`
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
	Method          types.String `tfsdk:"method"`
	ExpectedStatus  types.Int64  `tfsdk:"expected_status"`
	Interval        types.Int64  `tfsdk:"interval"`
	Timeout         types.Int64  `tfsdk:"timeout"`
	Regions         types.List   `tfsdk:"regions"`
	AlertChannelIDs types.List   `tfsdk:"alert_channel_ids"`
	Tags            types.List   `tfsdk:"tags"`
	Paused          types.Bool   `tfsdk:"paused"`
	Status          types.String `tfsdk:"status"`
}

func NewMonitorResource() resource.Resource {
	return &MonitorResource{}
}

func (r *MonitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (r *MonitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a Spork uptime monitor.",
		MarkdownDescription: "Manages a [Spork](https://sporkops.com) uptime monitor that periodically checks a target URL and reports its status.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier of the monitor.",
				MarkdownDescription: "The unique identifier of the monitor.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"target": schema.StringAttribute{
				Required:            true,
				Description:         "The URL to monitor. Must start with http:// or https://.",
				MarkdownDescription: "The URL to monitor. Must start with `http://` or `https://`.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^https?://`),
						"must start with http:// or https://",
					),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A human-readable name for the monitor (1-100 characters).",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("http"),
				Description:         "Monitor type. Default: http.",
				MarkdownDescription: "Monitor type. One of: `http`, `ssl`, `dns`, `keyword`, `tcp`, `ping`. Default: `http`.",
				Validators: []validator.String{
					stringvalidator.OneOf("http", "ssl", "dns", "keyword", "tcp", "ping"),
				},
			},
			"method": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("GET"),
				Description:         "HTTP method to use for checks. Default: GET.",
				MarkdownDescription: "HTTP method to use for checks. One of: `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `PATCH`, `OPTIONS`. Default: `GET`.",
				Validators: []validator.String{
					stringvalidator.OneOf("GET", "HEAD", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"),
				},
			},
			"expected_status": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(200),
				Description:         "Expected HTTP status code. Default: 200.",
				MarkdownDescription: "Expected HTTP status code (100-599). Default: `200`.",
				Validators: []validator.Int64{
					int64validator.Between(100, 599),
				},
			},
			"interval": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
				Description:         "Check interval in seconds (60-3600). Default: 60.",
				MarkdownDescription: "Check interval in seconds (`60`-`3600`). Default: `60`.",
				Validators: []validator.Int64{
					int64validator.Between(60, 3600),
				},
			},
			"timeout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(30),
				Description:         "Timeout in seconds for each check (5-120). Default: 30.",
				MarkdownDescription: "Timeout in seconds for each check (`5`-`120`). Default: `30`.",
				Validators: []validator.Int64{
					int64validator.Between(5, 120),
				},
			},
			"regions": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Description:         "Regions to check from. Default: [\"us-central1\"].",
				MarkdownDescription: "Regions to check from. Available: `us-central1`, `europe-west1`. Default: `[\"us-central1\"]`.",
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf("us-central1", "europe-west1"),
					),
				},
			},
			"alert_channel_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "IDs of alert channels to notify on status changes.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Tags for organizing monitors.",
			},
			"paused": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether the monitor is paused. Default: false.",
				MarkdownDescription: "Whether the monitor is paused. Default: `false`.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				Description:         "Current monitor status: up, down, degraded, paused, or pending.",
				MarkdownDescription: "Current monitor status: `up`, `down`, `degraded`, `paused`, or `pending`.",
			},
		},
	}
}

func (r *MonitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*SporkClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *SporkClient, got something else. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = client
}

func (r *MonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiMonitor := monitorFromModel(plan)

	result, err := r.client.CreateMonitor(ctx, apiMonitor)
	if err != nil {
		resp.Diagnostics.AddError("Error creating monitor", err.Error())
		return
	}

	state := monitorToModel(*result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetMonitor(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading monitor", err.Error())
		return
	}

	newState := monitorToModel(*result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *MonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiMonitor := monitorFromModel(plan)

	result, err := r.client.UpdateMonitor(ctx, state.ID.ValueString(), apiMonitor)
	if err != nil {
		resp.Diagnostics.AddError("Error updating monitor", err.Error())
		return
	}

	newState := monitorToModel(*result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *MonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting monitor", err.Error())
	}
}

func (r *MonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Conversion helpers

func monitorFromModel(model MonitorResourceModel) Monitor {
	mon := Monitor{
		Target:         model.Target.ValueString(),
		Name:           model.Name.ValueString(),
		Type:           model.Type.ValueString(),
		Method:         model.Method.ValueString(),
		ExpectedStatus: int(model.ExpectedStatus.ValueInt64()),
		Interval:       int(model.Interval.ValueInt64()),
		Timeout:        int(model.Timeout.ValueInt64()),
		Paused:         model.Paused.ValueBool(),
	}

	if !model.Regions.IsNull() && !model.Regions.IsUnknown() {
		var regions []string
		model.Regions.ElementsAs(context.Background(), &regions, false)
		mon.Regions = regions
	}

	if !model.AlertChannelIDs.IsNull() && !model.AlertChannelIDs.IsUnknown() {
		var ids []string
		model.AlertChannelIDs.ElementsAs(context.Background(), &ids, false)
		mon.AlertChannelIDs = ids
	}

	if !model.Tags.IsNull() && !model.Tags.IsUnknown() {
		var tags []string
		model.Tags.ElementsAs(context.Background(), &tags, false)
		mon.Tags = tags
	}

	return mon
}

func monitorToModel(m Monitor) MonitorResourceModel {
	regions, _ := types.ListValueFrom(context.Background(), types.StringType, m.Regions)
	if m.Regions == nil {
		regions, _ = types.ListValueFrom(context.Background(), types.StringType, []string{"us-central1"})
	}

	alertChannelIDs := types.ListNull(types.StringType)
	if m.AlertChannelIDs != nil {
		alertChannelIDs, _ = types.ListValueFrom(context.Background(), types.StringType, m.AlertChannelIDs)
	}

	tags := types.ListNull(types.StringType)
	if m.Tags != nil {
		tags, _ = types.ListValueFrom(context.Background(), types.StringType, m.Tags)
	}

	return MonitorResourceModel{
		ID:              types.StringValue(m.ID),
		Target:          types.StringValue(m.Target),
		Name:            types.StringValue(m.Name),
		Type:            types.StringValue(m.Type),
		Method:          types.StringValue(m.Method),
		ExpectedStatus:  types.Int64Value(int64(m.ExpectedStatus)),
		Interval:        types.Int64Value(int64(m.Interval)),
		Timeout:         types.Int64Value(int64(m.Timeout)),
		Regions:         regions,
		AlertChannelIDs: alertChannelIDs,
		Tags:            tags,
		Paused:          types.BoolValue(m.Paused),
		Status:          types.StringValue(m.Status),
	}
}
