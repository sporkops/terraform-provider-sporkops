package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sporkops/spork-go"
)

var _ datasource.DataSource = &StatusPageDataSource{}
var _ datasource.DataSourceWithConfigure = &StatusPageDataSource{}

type StatusPageDataSource struct {
	client *spork.Client
}

type StatusPageDataSourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Slug                    types.String `tfsdk:"slug"`
	Components              types.List   `tfsdk:"components"`
	ComponentGroups         types.List   `tfsdk:"component_groups"`
	CustomDomain            types.String `tfsdk:"custom_domain"`
	DomainStatus            types.String `tfsdk:"domain_status"`
	Theme                   types.String `tfsdk:"theme"`
	AccentColor             types.String `tfsdk:"accent_color"`
	FontFamily              types.String `tfsdk:"font_family"`
	HeaderStyle             types.String `tfsdk:"header_style"`
	LogoURL                 types.String `tfsdk:"logo_url"`
	WebhookURL              types.String `tfsdk:"webhook_url"`
	EmailSubscribersEnabled types.Bool   `tfsdk:"email_subscribers_enabled"`
	IsPublic                types.Bool   `tfsdk:"is_public"`
	CreatedAt               types.String `tfsdk:"created_at"`
	UpdatedAt               types.String `tfsdk:"updated_at"`
}

func NewStatusPageDataSource() datasource.DataSource {
	return &StatusPageDataSource{}
}

func (d *StatusPageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

func (d *StatusPageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches a Spork status page by ID or name.",
		MarkdownDescription: "Fetches a [Spork](https://sporkops.com) status page by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The unique identifier of the status page. Specify either id or name.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the status page. Specify either id or name.",
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
						"group": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the component group this component belongs to.",
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
			"font_family": schema.StringAttribute{
				Computed:    true,
				Description: "Font family for the status page.",
			},
			"header_style": schema.StringAttribute{
				Computed:    true,
				Description: "Header style for the status page.",
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
	}
}

func (d *StatusPageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StatusPageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config StatusPageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result *spork.StatusPage

	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		r, err := d.client.GetStatusPage(ctx, config.ID.ValueString())
		if err != nil {
			if spork.IsNotFound(err) {
				resp.Diagnostics.AddError("Status Page Not Found", "No status page found with ID: "+config.ID.ValueString())
				return
			}
			resp.Diagnostics.AddError("Error reading status page", err.Error())
			return
		}
		result = r
	} else if !config.Name.IsNull() && config.Name.ValueString() != "" {
		pages, err := d.client.ListStatusPages(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Error listing status pages", err.Error())
			return
		}
		var matches []spork.StatusPage
		for _, p := range pages {
			if p.Name == config.Name.ValueString() {
				matches = append(matches, p)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("Status Page Not Found", "No status page found with name: "+config.Name.ValueString())
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Status Pages Found",
				fmt.Sprintf("Found %d status pages with name %q. Use id to specify the exact status page.", len(matches), config.Name.ValueString()),
			)
			return
		}
		result = &matches[0]
	} else {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Specify either id or name to look up a status page.",
		)
		return
	}

	state := statusPageToDataSourceModel(*result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func statusPageToDataSourceModel(p spork.StatusPage) StatusPageDataSourceModel {
	model := StatusPageDataSourceModel{
		ID:                      types.StringValue(p.ID),
		Name:                    types.StringValue(p.Name),
		Slug:                    types.StringValue(p.Slug),
		Theme:                   types.StringValue(p.Theme),
		IsPublic:                types.BoolValue(p.IsPublic),
		EmailSubscribersEnabled: types.BoolValue(p.EmailSubscribersEnabled),
		CreatedAt:               types.StringValue(p.CreatedAt),
		UpdatedAt:               types.StringValue(p.UpdatedAt),
	}

	if p.CustomDomain != "" {
		model.CustomDomain = types.StringValue(p.CustomDomain)
		model.DomainStatus = types.StringValue(p.DomainStatus)
	} else {
		model.CustomDomain = types.StringNull()
		model.DomainStatus = types.StringNull()
	}

	if p.AccentColor != "" {
		model.AccentColor = types.StringValue(p.AccentColor)
	} else {
		model.AccentColor = types.StringNull()
	}

	model.FontFamily = types.StringValue(p.FontFamily)
	model.HeaderStyle = types.StringValue(p.HeaderStyle)

	if p.LogoURL != "" {
		model.LogoURL = types.StringValue(p.LogoURL)
	} else {
		model.LogoURL = types.StringNull()
	}

	if p.WebhookURL != "" {
		model.WebhookURL = types.StringValue(p.WebhookURL)
	} else {
		model.WebhookURL = types.StringNull()
	}

	// Build group ID -> name lookup from component groups
	groupIDToName := make(map[string]string, len(p.ComponentGroups))
	for _, g := range p.ComponentGroups {
		groupIDToName[g.ID] = g.Name
	}

	// Components
	if len(p.Components) > 0 {
		var compValues []attr.Value
		for _, c := range p.Components {
			desc := types.StringNull()
			if c.Description != "" {
				desc = types.StringValue(c.Description)
			}
			groupName := types.StringNull()
			if c.GroupID != "" {
				if name, ok := groupIDToName[c.GroupID]; ok {
					groupName = types.StringValue(name)
				}
			}
			compValues = append(compValues, types.ObjectValueMust(componentAttrTypes, map[string]attr.Value{
				"id":           types.StringValue(c.ID),
				"monitor_id":   types.StringValue(c.MonitorID),
				"display_name": types.StringValue(c.DisplayName),
				"description":  desc,
				"group":        groupName,
				"order":        types.Int64Value(int64(c.Order)),
			}))
		}
		model.Components = types.ListValueMust(types.ObjectType{AttrTypes: componentAttrTypes}, compValues)
	} else {
		model.Components = types.ListNull(types.ObjectType{AttrTypes: componentAttrTypes})
	}

	// Component groups
	if len(p.ComponentGroups) > 0 {
		var groupValues []attr.Value
		for _, g := range p.ComponentGroups {
			desc := types.StringNull()
			if g.Description != "" {
				desc = types.StringValue(g.Description)
			}
			groupValues = append(groupValues, types.ObjectValueMust(componentGroupAttrTypes, map[string]attr.Value{
				"id":          types.StringValue(g.ID),
				"name":        types.StringValue(g.Name),
				"description": desc,
				"order":       types.Int64Value(int64(g.Order)),
			}))
		}
		model.ComponentGroups = types.ListValueMust(types.ObjectType{AttrTypes: componentGroupAttrTypes}, groupValues)
	} else {
		model.ComponentGroups = types.ListNull(types.ObjectType{AttrTypes: componentGroupAttrTypes})
	}

	return model
}
