package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sporkops/spork-go"
)

var _ datasource.DataSource = &AlertChannelsDataSource{}
var _ datasource.DataSourceWithConfigure = &AlertChannelsDataSource{}

type AlertChannelsDataSource struct {
	client *spork.Client
}

type AlertChannelsDataSourceModel struct {
	Type          types.String                   `tfsdk:"type"`
	AlertChannels []AlertChannelDataSourceModel  `tfsdk:"alert_channels"`
}

func NewAlertChannelsDataSource() datasource.DataSource {
	return &AlertChannelsDataSource{}
}

func (d *AlertChannelsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_channels"
}

func (d *AlertChannelsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches all Spork alert channels with optional filtering.",
		MarkdownDescription: "Fetches all [Spork](https://sporkops.com) alert channels with optional filtering by type.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Optional:            true,
				Description:         "Filter alert channels by type.",
				MarkdownDescription: "Filter alert channels by type. One of: `email`, `webhook`, `slack`, `discord`, `teams`, `pagerduty`, `telegram`, `googlechat`.",
			},
			"alert_channels": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of alert channels matching the filters.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the alert channel.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the alert channel.",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							Description:         "The channel type.",
							MarkdownDescription: "The channel type: `email`, `webhook`, `slack`, `discord`, `teams`, `pagerduty`, `telegram`, or `googlechat`.",
						},
						"config": schema.MapAttribute{
							Computed:    true,
							Sensitive:   true,
							ElementType: types.StringType,
							Description: "Channel-specific configuration as key-value pairs. Marked sensitive because it contains webhook URLs, integration keys, and bot tokens.",
						},
						"verified": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the channel has been verified. Relevant for email type.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the alert channel was created.",
						},
					},
				},
			},
		},
	}
}

func (d *AlertChannelsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AlertChannelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AlertChannelsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	channels, err := d.client.ListAlertChannels(ctx)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Error listing alert channels", err)
		return
	}

	// Apply optional type filter
	filterType := ""
	if !config.Type.IsNull() && !config.Type.IsUnknown() {
		filterType = config.Type.ValueString()
	}

	var results []AlertChannelDataSourceModel
	for _, c := range channels {
		if filterType != "" && c.Type != filterType {
			continue
		}

		var convDiags diag.Diagnostics
		configMap := types.MapNull(types.StringType)
		if c.Config != nil {
			configMap, convDiags = types.MapValueFrom(ctx, types.StringType, c.Config)
			resp.Diagnostics.Append(convDiags...)
		}

		results = append(results, AlertChannelDataSourceModel{
			ID:        types.StringValue(c.ID),
			Name:      types.StringValue(c.Name),
			Type:      types.StringValue(c.Type),
			Config:    configMap,
			Verified:  types.BoolValue(c.Verified),
			CreatedAt: types.StringValue(c.CreatedAt),
		})
	}

	if results == nil {
		results = []AlertChannelDataSourceModel{}
	}

	state := AlertChannelsDataSourceModel{
		Type:          config.Type,
		AlertChannels: results,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
