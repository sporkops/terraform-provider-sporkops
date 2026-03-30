package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &StatusPagesDataSource{}
var _ datasource.DataSourceWithConfigure = &StatusPagesDataSource{}

type StatusPagesDataSource struct {
	client *SporkClient
}

type StatusPagesDataSourceModel struct {
	StatusPages []StatusPageDataSourceModel `tfsdk:"status_pages"`
}

func NewStatusPagesDataSource() datasource.DataSource {
	return &StatusPagesDataSource{}
}

func (d *StatusPagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_pages"
}

func (d *StatusPagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches all Spork status pages.",
		MarkdownDescription: "Fetches all [Spork](https://sporkops.com) status pages.",
		Attributes: map[string]schema.Attribute{
			"status_pages": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of status pages.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the status page.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the status page.",
						},
						"slug": schema.StringAttribute{
							Computed:    true,
							Description: "URL-safe slug for the status page.",
						},
						"components": schema.ListNestedAttribute{
							Computed:    true,
							Description: "Components displayed on the status page.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Computed:    true,
										Description: "The unique identifier of the component.",
									},
									"monitor_id": schema.StringAttribute{
										Computed:    true,
										Description: "The ID of the monitor displayed on the status page.",
									},
									"display_name": schema.StringAttribute{
										Computed:    true,
										Description: "The display name shown on the status page.",
									},
									"description": schema.StringAttribute{
										Computed:    true,
										Description: "A description of the component.",
									},
									"group_id": schema.StringAttribute{
										Computed:    true,
										Description: "The ID of the component group this component belongs to.",
									},
									"order": schema.Int64Attribute{
										Computed:    true,
										Description: "Display order of the component.",
									},
								},
							},
						},
						"component_groups": schema.ListNestedAttribute{
							Computed:    true,
							Description: "Component groups for organizing components into named sections.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Computed:    true,
										Description: "The unique identifier of the component group.",
									},
									"name": schema.StringAttribute{
										Computed:    true,
										Description: "The display name of the component group.",
									},
									"description": schema.StringAttribute{
										Computed:    true,
										Description: "A description of the component group.",
									},
									"order": schema.Int64Attribute{
										Computed:    true,
										Description: "Display order of the component group.",
									},
								},
							},
						},
						"custom_domain": schema.StringAttribute{
							Computed:    true,
							Description: "Custom domain for the status page.",
						},
						"domain_status": schema.StringAttribute{
							Computed:    true,
							Description: "Verification status of the custom domain.",
						},
						"theme": schema.StringAttribute{
							Computed:    true,
							Description: "Color theme for the status page.",
						},
						"accent_color": schema.StringAttribute{
							Computed:    true,
							Description: "Accent color as a hex color.",
						},
						"logo_url": schema.StringAttribute{
							Computed:    true,
							Description: "URL of the logo displayed on the status page.",
						},
						"webhook_url": schema.StringAttribute{
							Computed:    true,
							Description: "Webhook URL for incident notifications.",
						},
						"email_subscribers_enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether email subscriber notifications are enabled.",
						},
						"font_family": schema.StringAttribute{
							Computed:    true,
							Description: "Font family for the status page.",
						},
						"header_style": schema.StringAttribute{
							Computed:    true,
							Description: "Header style for the status page.",
						},
						"is_public": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the status page is publicly accessible.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the status page was created.",
						},
						"updated_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the status page was last updated.",
						},
					},
				},
			},
		},
	}
}

func (d *StatusPagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StatusPagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	pages, err := d.client.ListStatusPages(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing status pages", err.Error())
		return
	}

	var results []StatusPageDataSourceModel
	for _, p := range pages {
		results = append(results, statusPageToDataSourceModel(p))
	}

	if results == nil {
		results = []StatusPageDataSourceModel{}
	}

	state := StatusPagesDataSourceModel{
		StatusPages: results,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
