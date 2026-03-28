package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &StatusPageDataSource{}
var _ datasource.DataSourceWithConfigure = &StatusPageDataSource{}

type StatusPageDataSource struct {
	client *SporkClient
}

type StatusPageDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Slug         types.String `tfsdk:"slug"`
	Components   types.List   `tfsdk:"components"`
	CustomDomain types.String `tfsdk:"custom_domain"`
	DomainStatus types.String `tfsdk:"domain_status"`
	Theme        types.String `tfsdk:"theme"`
	AccentColor  types.String `tfsdk:"accent_color"`
	LogoURL      types.String `tfsdk:"logo_url"`
	IsPublic     types.Bool   `tfsdk:"is_public"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
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
						"order": schema.Int64Attribute{
							Computed:    true,
							Description: "Display order of the component.",
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

func (d *StatusPageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config StatusPageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result *StatusPage

	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		r, err := d.client.GetStatusPage(ctx, config.ID.ValueString())
		if err != nil {
			if errors.Is(err, ErrNotFound) {
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
		var matches []StatusPage
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

	state := statusPageToDataSourceModel(ctx, *result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func statusPageToDataSourceModel(ctx context.Context, p StatusPage) StatusPageDataSourceModel {
	model := StatusPageDataSourceModel{
		ID:        types.StringValue(p.ID),
		Name:      types.StringValue(p.Name),
		Slug:      types.StringValue(p.Slug),
		Theme:     types.StringValue(p.Theme),
		IsPublic:  types.BoolValue(p.IsPublic),
		CreatedAt: types.StringValue(p.CreatedAt),
		UpdatedAt: types.StringValue(p.UpdatedAt),
	}

	if p.CustomDomain != "" {
		model.CustomDomain = types.StringValue(p.CustomDomain)
		model.DomainStatus = types.StringValue(p.DomainStatus)
	} else {
		model.CustomDomain = types.StringValue("")
		model.DomainStatus = types.StringValue("")
	}

	if p.AccentColor != "" {
		model.AccentColor = types.StringValue(p.AccentColor)
	} else {
		model.AccentColor = types.StringValue("")
	}

	if p.LogoURL != "" {
		model.LogoURL = types.StringValue(p.LogoURL)
	} else {
		model.LogoURL = types.StringValue("")
	}

	if len(p.Components) > 0 {
		var compValues []attr.Value
		for _, c := range p.Components {
			desc := types.StringValue("")
			if c.Description != "" {
				desc = types.StringValue(c.Description)
			}
			compObj, _ := types.ObjectValue(componentAttrTypes, map[string]attr.Value{
				"id":           types.StringValue(c.ID),
				"monitor_id":   types.StringValue(c.MonitorID),
				"display_name": types.StringValue(c.DisplayName),
				"description":  desc,
				"order":        types.Int64Value(int64(c.Order)),
			})
			compValues = append(compValues, compObj)
		}
		compList, _ := types.ListValue(types.ObjectType{AttrTypes: componentAttrTypes}, compValues)
		model.Components = compList
	} else {
		model.Components = types.ListValueMust(types.ObjectType{AttrTypes: componentAttrTypes}, []attr.Value{})
	}

	return model
}
