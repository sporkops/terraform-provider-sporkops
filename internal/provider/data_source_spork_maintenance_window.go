package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sporkops/spork-go"
)

var _ datasource.DataSource = &MaintenanceWindowDataSource{}
var _ datasource.DataSourceWithConfigure = &MaintenanceWindowDataSource{}

// MaintenanceWindowDataSource looks up a single maintenance window by
// ID or by name. Mirrors the alert-channel data source's contract so
// authors can reuse the same idioms they already know.
type MaintenanceWindowDataSource struct {
	client *spork.Client
}

type MaintenanceWindowDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	MonitorIDs        types.Set    `tfsdk:"monitor_ids"`
	TagSelectors      types.Set    `tfsdk:"tag_selectors"`
	AllMonitors       types.Bool   `tfsdk:"all_monitors"`
	Timezone          types.String `tfsdk:"timezone"`
	StartAt           types.String `tfsdk:"start_at"`
	EndAt             types.String `tfsdk:"end_at"`
	RecurrenceType    types.String `tfsdk:"recurrence_type"`
	RecurrenceDays    types.List   `tfsdk:"recurrence_days"`
	RecurrenceUntil   types.String `tfsdk:"recurrence_until"`
	SuppressAlerts    types.Bool   `tfsdk:"suppress_alerts"`
	ExcludeFromUptime types.Bool   `tfsdk:"exclude_from_uptime"`
	PauseChecks       types.Bool   `tfsdk:"pause_checks"`
	State             types.String `tfsdk:"state"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

func NewMaintenanceWindowDataSource() datasource.DataSource {
	return &MaintenanceWindowDataSource{}
}

func (d *MaintenanceWindowDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_window"
}

func (d *MaintenanceWindowDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches a Spork maintenance window by ID or name.",
		MarkdownDescription: "Fetches a [Spork](https://sporkops.com) maintenance window by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The unique identifier of the maintenance window. Specify either id or name.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the maintenance window. Specify either id or name.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Optional description displayed alongside the window.",
			},
			"monitor_ids": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Monitor IDs the window applies to.",
			},
			"tag_selectors": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Monitor tags the window applies to (OR semantics).",
			},
			"all_monitors": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the window applies to every monitor.",
			},
			"timezone": schema.StringAttribute{
				Computed:    true,
				Description: "IANA timezone name.",
			},
			"start_at": schema.StringAttribute{
				Computed:    true,
				Description: "Window start (RFC3339 UTC).",
			},
			"end_at": schema.StringAttribute{
				Computed:    true,
				Description: "Window end (RFC3339 UTC).",
			},
			"recurrence_type": schema.StringAttribute{
				Computed:    true,
				Description: "Recurrence type: daily, weekly, or monthly. Empty for one-time windows.",
			},
			"recurrence_days": schema.ListAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
				Description: "Days of the week (0-6, Sun=0) or month (1-31).",
			},
			"recurrence_until": schema.StringAttribute{
				Computed:    true,
				Description: "Cap on the recurring series (RFC3339 UTC).",
			},
			"suppress_alerts": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether alerts are suppressed during the window.",
			},
			"exclude_from_uptime": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether in-window checks are excluded from uptime calculations.",
			},
			"pause_checks": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether checks are paused entirely during the window.",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "Computed lifecycle state: scheduled, in_progress, completed, or cancelled.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the window was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the window was last updated.",
			},
		},
	}
}

func (d *MaintenanceWindowDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*spork.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *spork.Client, got something else. Please report this issue to the provider developers.",
		)
		return
	}
	d.client = client
}

func (d *MaintenanceWindowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MaintenanceWindowDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result *spork.MaintenanceWindow
	switch {
	case !config.ID.IsNull() && config.ID.ValueString() != "":
		r, err := d.client.GetMaintenanceWindow(ctx, config.ID.ValueString())
		if err != nil {
			if spork.IsNotFound(err) {
				resp.Diagnostics.AddError("Maintenance Window Not Found", "No maintenance window found with ID: "+config.ID.ValueString())
				return
			}
			addAPIError(&resp.Diagnostics, "Error reading maintenance window", err)
			return
		}
		result = r
	case !config.Name.IsNull() && config.Name.ValueString() != "":
		windows, err := d.client.ListMaintenanceWindows(ctx)
		if err != nil {
			addAPIError(&resp.Diagnostics, "Error listing maintenance windows", err)
			return
		}
		var matches []spork.MaintenanceWindow
		for _, w := range windows {
			if w.Name == config.Name.ValueString() {
				matches = append(matches, w)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("Maintenance Window Not Found", "No maintenance window found with name: "+config.Name.ValueString())
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Maintenance Windows Found",
				fmt.Sprintf("Found %d maintenance windows with name %q. Use id to specify the exact window.", len(matches), config.Name.ValueString()),
			)
			return
		}
		result = &matches[0]
	default:
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Specify either id or name to look up a maintenance window.",
		)
		return
	}

	state := maintenanceWindowDataModelFromAPI(ctx, *result, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// maintenanceWindowDataModelFromAPI converts a spork.MaintenanceWindow to
// the data-source model. Shared with the plural data source below.
func maintenanceWindowDataModelFromAPI(ctx context.Context, w spork.MaintenanceWindow, diags *diag.Diagnostics) MaintenanceWindowDataSourceModel {
	m := MaintenanceWindowDataSourceModel{
		ID:              types.StringValue(w.ID),
		Name:            types.StringValue(w.Name),
		Description:     types.StringValue(w.Description),
		Timezone:        types.StringValue(w.Timezone),
		StartAt:         types.StringValue(w.StartAt),
		EndAt:           types.StringValue(w.EndAt),
		RecurrenceType:  types.StringValue(w.RecurrenceType),
		RecurrenceUntil: types.StringValue(w.RecurrenceUntil),
		State:           types.StringValue(w.State),
		CreatedAt:       types.StringValue(w.CreatedAt),
		UpdatedAt:       types.StringValue(w.UpdatedAt),
	}

	if len(w.MonitorIDs) > 0 {
		set, d := types.SetValueFrom(ctx, types.StringType, w.MonitorIDs)
		diags.Append(d...)
		m.MonitorIDs = set
	} else {
		m.MonitorIDs = types.SetNull(types.StringType)
	}
	if len(w.TagSelectors) > 0 {
		set, d := types.SetValueFrom(ctx, types.StringType, w.TagSelectors)
		diags.Append(d...)
		m.TagSelectors = set
	} else {
		m.TagSelectors = types.SetNull(types.StringType)
	}
	if len(w.RecurrenceDays) > 0 {
		days := make([]int64, len(w.RecurrenceDays))
		for i, dv := range w.RecurrenceDays {
			days[i] = int64(dv)
		}
		list, d := types.ListValueFrom(ctx, types.Int64Type, days)
		diags.Append(d...)
		m.RecurrenceDays = list
	} else {
		m.RecurrenceDays = types.ListNull(types.Int64Type)
	}

	if w.AllMonitors != nil {
		m.AllMonitors = types.BoolValue(*w.AllMonitors)
	} else {
		m.AllMonitors = types.BoolValue(false)
	}
	if w.SuppressAlerts != nil {
		m.SuppressAlerts = types.BoolValue(*w.SuppressAlerts)
	} else {
		m.SuppressAlerts = types.BoolValue(true)
	}
	if w.ExcludeFromUptime != nil {
		m.ExcludeFromUptime = types.BoolValue(*w.ExcludeFromUptime)
	} else {
		m.ExcludeFromUptime = types.BoolValue(true)
	}
	if w.PauseChecks != nil {
		m.PauseChecks = types.BoolValue(*w.PauseChecks)
	} else {
		m.PauseChecks = types.BoolValue(false)
	}

	return m
}
