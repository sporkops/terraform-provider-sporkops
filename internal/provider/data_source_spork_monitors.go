package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sporkops/spork-go"
)

var _ datasource.DataSource = &MonitorsDataSource{}
var _ datasource.DataSourceWithConfigure = &MonitorsDataSource{}

type MonitorsDataSource struct {
	client *spork.Client
}

type MonitorsDataSourceModel struct {
	Type     types.String              `tfsdk:"type"`
	Status   types.String              `tfsdk:"status"`
	Monitors []MonitorDataSourceModel  `tfsdk:"monitors"`
}

func NewMonitorsDataSource() datasource.DataSource {
	return &MonitorsDataSource{}
}

func (d *MonitorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitors"
}

func (d *MonitorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches all Spork uptime monitors with optional filtering.",
		MarkdownDescription: "Fetches all [Spork](https://sporkops.com) uptime monitors with optional filtering by type or status.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Optional:            true,
				Description:         "Filter monitors by type.",
				MarkdownDescription: "Filter monitors by type. One of: `http`, `ssl`, `dns`, `keyword`, `tcp`, `ping`.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				Description:         "Filter monitors by status.",
				MarkdownDescription: "Filter monitors by status. One of: `up`, `down`, `degraded`, `paused`, `pending`.",
			},
			"monitors": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of monitors matching the filters.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the monitor.",
						},
						"target": schema.StringAttribute{
							Computed:    true,
							Description: "The URL being monitored.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the monitor.",
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
				},
			},
		},
	}
}

func (d *MonitorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MonitorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MonitorsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	monitors, err := d.client.ListMonitors(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing monitors", err.Error())
		return
	}

	// Apply optional filters
	filterType := ""
	if !config.Type.IsNull() && !config.Type.IsUnknown() {
		filterType = config.Type.ValueString()
	}
	filterStatus := ""
	if !config.Status.IsNull() && !config.Status.IsUnknown() {
		filterStatus = config.Status.ValueString()
	}

	var results []MonitorDataSourceModel
	for _, m := range monitors {
		if filterType != "" && m.Type != filterType {
			continue
		}
		if filterStatus != "" && m.Status != filterStatus {
			continue
		}

		regions, _ := types.ListValueFrom(ctx, types.StringType, m.Regions)
		if m.Regions == nil {
			regions, _ = types.ListValueFrom(ctx, types.StringType, []string{})
		}

		alertChannelIDs := types.ListNull(types.StringType)
		if m.AlertChannelIDs != nil {
			alertChannelIDs, _ = types.ListValueFrom(ctx, types.StringType, m.AlertChannelIDs)
		}

		tags := types.ListNull(types.StringType)
		if m.Tags != nil {
			tags, _ = types.ListValueFrom(ctx, types.StringType, m.Tags)
		}

		headers := types.MapNull(types.StringType)
		if m.Headers != nil {
			headers, _ = types.MapValueFrom(ctx, types.StringType, m.Headers)
		}

		results = append(results, MonitorDataSourceModel{
			ID:              types.StringValue(m.ID),
			Target:          types.StringValue(m.Target),
			Name:            types.StringValue(m.Name),
			Type:            types.StringValue(m.Type),
			Method:          types.StringValue(m.Method),
			Interval:        types.Int64Value(int64(m.Interval)),
			Timeout:         types.Int64Value(int64(m.Timeout)),
			ExpectedStatus:  types.Int64Value(int64(m.ExpectedStatus)),
			Paused:          types.BoolValue(m.Paused != nil && *m.Paused),
			Status:          types.StringValue(m.Status),
			Regions:         regions,
			AlertChannelIDs: alertChannelIDs,
			Tags:            tags,
			Headers:         headers,
			Body:            types.StringValue(m.Body),
			Keyword:         types.StringValue(m.Keyword),
			KeywordType:     types.StringValue(m.KeywordType),
			SSLWarnDays:     types.Int64Value(int64(m.SSLWarnDays)),
			CreatedAt:       types.StringValue(m.CreatedAt),
			UpdatedAt:       types.StringValue(m.UpdatedAt),
		})
	}

	if results == nil {
		results = []MonitorDataSourceModel{}
	}

	state := MonitorsDataSourceModel{
		Type:     config.Type,
		Status:   config.Status,
		Monitors: results,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
