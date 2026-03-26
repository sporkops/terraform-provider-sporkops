package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AlertChannelDataSource{}
var _ datasource.DataSourceWithConfigure = &AlertChannelDataSource{}

type AlertChannelDataSource struct {
	client *SporkClient
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
		Description:         "Fetches a Spork alert channel by ID.",
		MarkdownDescription: "Fetches a [Spork](https://sporkops.com) alert channel by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the alert channel.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "A friendly name for the alert channel.",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				Description:         "The channel type.",
				MarkdownDescription: "The channel type: `email`, `webhook`, `slack`, `discord`, `teams`, `pagerduty`, `telegram`, or `googlechat`.",
			},
			"config": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Channel-specific configuration as key-value pairs.",
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

	client, ok := req.ProviderData.(*SporkClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *SporkClient, got something else. Please report this issue to the provider developers.",
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

	result, err := d.client.GetAlertChannel(ctx, config.ID.ValueString())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			resp.Diagnostics.AddError(
				"Alert Channel Not Found",
				"No alert channel found with ID: "+config.ID.ValueString(),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading alert channel", err.Error())
		return
	}

	configMap := types.MapNull(types.StringType)
	if result.Config != nil {
		configMap, _ = types.MapValueFrom(ctx, types.StringType, result.Config)
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
