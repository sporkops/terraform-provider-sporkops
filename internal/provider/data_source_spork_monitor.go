package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sporkops/spork-go"
)

var _ datasource.DataSource = &MonitorDataSource{}
var _ datasource.DataSourceWithConfigure = &MonitorDataSource{}

type MonitorDataSource struct {
	client *spork.Client
}

type MonitorDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Target           types.String `tfsdk:"target"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	Method           types.String `tfsdk:"method"`
	Interval         types.Int64  `tfsdk:"interval"`
	Timeout          types.Int64  `tfsdk:"timeout"`
	ExpectedStatus   types.Int64  `tfsdk:"expected_status"`
	Paused           types.Bool   `tfsdk:"paused"`
	Status           types.String `tfsdk:"status"`
	Regions          types.List   `tfsdk:"regions"`
	AlertChannelIDs  types.List   `tfsdk:"alert_channel_ids"`
	Tags             types.List   `tfsdk:"tags"`
	Headers          types.Map    `tfsdk:"headers"`
	Body             types.String `tfsdk:"body"`
	Keyword          types.String `tfsdk:"keyword"`
	KeywordType      types.String `tfsdk:"keyword_type"`
	SSLWarnDays      types.Int64  `tfsdk:"ssl_warn_days"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func NewMonitorDataSource() datasource.DataSource {
	return &MonitorDataSource{}
}

func (d *MonitorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (d *MonitorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches a Spork uptime monitor by ID or name.",
		MarkdownDescription: "Fetches a [Spork](https://sporkops.com) uptime monitor by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The unique identifier of the monitor. Specify either id or name.",
			},
			"target": schema.StringAttribute{
				Computed:    true,
				Description: "The URL being monitored.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the monitor. Specify either id or name.",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				Description:         "Monitor type.",
				MarkdownDescription: "Monitor type. One of: `http`, `ssl`, `dns`, `keyword`, `tcp`, `ping`.",
			},
			"method": schema.StringAttribute{
				Computed:            true,
				Description:         "HTTP method used for checks.",
				MarkdownDescription: "HTTP method used for checks. One of: `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `PATCH`, `OPTIONS`.",
			},
			"interval": schema.Int64Attribute{
				Computed:    true,
				Description: "Check interval in seconds.",
			},
			"timeout": schema.Int64Attribute{
				Computed:    true,
				Description: "Timeout in seconds for each check.",
			},
			"expected_status": schema.Int64Attribute{
				Computed:    true,
				Description: "Expected HTTP status code.",
			},
			"paused": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the monitor is paused.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				Description:         "Current monitor status: up, down, degraded, paused, or pending.",
				MarkdownDescription: "Current monitor status: `up`, `down`, `degraded`, `paused`, or `pending`.",
			},
			"regions": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				Description:         "Regions the monitor checks from.",
				MarkdownDescription: "Regions the monitor checks from. Available: `us-central1`, `europe-west1`.",
			},
			"alert_channel_ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "IDs of alert channels notified on status changes.",
			},
			"tags": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags for organizing monitors.",
			},
			"headers": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Custom HTTP request headers sent with each check.",
			},
			"body": schema.StringAttribute{
				Computed:    true,
				Description: "HTTP request body sent with each check.",
			},
			"keyword": schema.StringAttribute{
				Computed:    true,
				Description: "The keyword searched for in the response body. Set when type is \"keyword\".",
			},
			"keyword_type": schema.StringAttribute{
				Computed:            true,
				Description:         "Whether the keyword must exist or not in the response.",
				MarkdownDescription: "Whether the keyword must exist or not in the response. One of: `exists`, `not_exists`.",
			},
			"ssl_warn_days": schema.Int64Attribute{
				Computed:    true,
				Description: "Days before SSL certificate expiry to trigger a warning. Set when type is \"ssl\".",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the monitor was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the monitor was last updated.",
			},
		},
	}
}

func (d *MonitorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MonitorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MonitorDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result *spork.Monitor

	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		// Lookup by ID
		r, err := d.client.GetMonitor(ctx, config.ID.ValueString())
		if err != nil {
			if spork.IsNotFound(err) {
				resp.Diagnostics.AddError("Monitor Not Found", "No monitor found with ID: "+config.ID.ValueString())
				return
			}
			resp.Diagnostics.AddError("Error reading monitor", err.Error())
			return
		}
		result = r
	} else if !config.Name.IsNull() && config.Name.ValueString() != "" {
		// Lookup by name
		monitors, err := d.client.ListMonitors(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Error listing monitors", err.Error())
			return
		}
		var matches []spork.Monitor
		for _, m := range monitors {
			if m.Name == config.Name.ValueString() {
				matches = append(matches, m)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("Monitor Not Found", "No monitor found with name: "+config.Name.ValueString())
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Monitors Found",
				fmt.Sprintf("Found %d monitors with name %q. Use id to specify the exact monitor.", len(matches), config.Name.ValueString()),
			)
			return
		}
		result = &matches[0]
	} else {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Specify either id or name to look up a monitor.",
		)
		return
	}

	regions, _ := types.ListValueFrom(ctx, types.StringType, result.Regions)
	if result.Regions == nil {
		regions, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}

	alertChannelIDs := types.ListNull(types.StringType)
	if result.AlertChannelIDs != nil {
		alertChannelIDs, _ = types.ListValueFrom(ctx, types.StringType, result.AlertChannelIDs)
	}

	tags := types.ListNull(types.StringType)
	if result.Tags != nil {
		tags, _ = types.ListValueFrom(ctx, types.StringType, result.Tags)
	}

	headers := types.MapNull(types.StringType)
	if result.Headers != nil {
		headers, _ = types.MapValueFrom(ctx, types.StringType, result.Headers)
	}

	state := MonitorDataSourceModel{
		ID:              types.StringValue(result.ID),
		Target:          types.StringValue(result.Target),
		Name:            types.StringValue(result.Name),
		Type:            types.StringValue(result.Type),
		Method:          types.StringValue(result.Method),
		Interval:        types.Int64Value(int64(result.Interval)),
		Timeout:         types.Int64Value(int64(result.Timeout)),
		ExpectedStatus:  types.Int64Value(int64(result.ExpectedStatus)),
		Paused:          types.BoolValue(result.Paused != nil && *result.Paused),
		Status:          types.StringValue(result.Status),
		Regions:         regions,
		AlertChannelIDs: alertChannelIDs,
		Tags:            tags,
		Headers:         headers,
		Body:            types.StringValue(result.Body),
		Keyword:         types.StringValue(result.Keyword),
		KeywordType:     types.StringValue(result.KeywordType),
		SSLWarnDays:     types.Int64Value(int64(result.SSLWarnDays)),
		CreatedAt:       types.StringValue(result.CreatedAt),
		UpdatedAt:       types.StringValue(result.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
