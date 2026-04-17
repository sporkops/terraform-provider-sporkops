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

var _ datasource.DataSource = &AlertChannelDataSource{}
var _ datasource.DataSourceWithConfigure = &AlertChannelDataSource{}

type AlertChannelDataSource struct {
	client *spork.Client
}

type AlertChannelDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Config    types.Map    `tfsdk:"config"`
	Verified  types.Bool   `tfsdk:"verified"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func NewAlertChannelDataSource() datasource.DataSource {
	return &AlertChannelDataSource{}
}

func (d *AlertChannelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_channel"
}

func (d *AlertChannelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches a Spork alert channel by ID or name.",
		MarkdownDescription: "Fetches a [Spork](https://sporkops.com) alert channel by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The unique identifier of the alert channel. Specify either id or name.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the alert channel. Specify either id or name.",
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
	}
}

func (d *AlertChannelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AlertChannelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AlertChannelDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result *spork.AlertChannel

	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		// Lookup by ID
		r, err := d.client.GetAlertChannel(ctx, config.ID.ValueString())
		if err != nil {
			if spork.IsNotFound(err) {
				resp.Diagnostics.AddError("Alert Channel Not Found", "No alert channel found with ID: "+config.ID.ValueString())
				return
			}
			addAPIError(&resp.Diagnostics, "Error reading alert channel", err)
			return
		}
		result = r
	} else if !config.Name.IsNull() && config.Name.ValueString() != "" {
		// Lookup by name
		channels, err := d.client.ListAlertChannels(ctx)
		if err != nil {
			addAPIError(&resp.Diagnostics, "Error listing alert channels", err)
			return
		}
		var matches []spork.AlertChannel
		for _, c := range channels {
			if c.Name == config.Name.ValueString() {
				matches = append(matches, c)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("Alert Channel Not Found", "No alert channel found with name: "+config.Name.ValueString())
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Alert Channels Found",
				fmt.Sprintf("Found %d alert channels with name %q. Use id to specify the exact alert channel.", len(matches), config.Name.ValueString()),
			)
			return
		}
		result = &matches[0]
	} else {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Specify either id or name to look up an alert channel.",
		)
		return
	}

	var convDiags diag.Diagnostics
	configMap := types.MapNull(types.StringType)
	if result.Config != nil {
		configMap, convDiags = types.MapValueFrom(ctx, types.StringType, result.Config)
		resp.Diagnostics.Append(convDiags...)
	}

	state := AlertChannelDataSourceModel{
		ID:        types.StringValue(result.ID),
		Name:      types.StringValue(result.Name),
		Type:      types.StringValue(result.Type),
		Config:    configMap,
		Verified:  types.BoolValue(result.Verified),
		CreatedAt: types.StringValue(result.CreatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
