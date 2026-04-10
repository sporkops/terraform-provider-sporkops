package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sporkops/spork-go"
)

var _ datasource.DataSource = &MembersDataSource{}
var _ datasource.DataSourceWithConfigure = &MembersDataSource{}

type MembersDataSource struct {
	client *spork.Client
}

type MembersDataSourceModel struct {
	Role    types.String             `tfsdk:"role"`
	Status  types.String             `tfsdk:"status"`
	Members []MemberDataSourceModel  `tfsdk:"members"`
}

type MemberDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func NewMembersDataSource() datasource.DataSource {
	return &MembersDataSource{}
}

func (d *MembersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_members"
}

func (d *MembersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches all Spork organization members with optional filtering.",
		MarkdownDescription: "Fetches all [Spork](https://sporkops.com) organization members with optional filtering by role or status.",
		Attributes: map[string]schema.Attribute{
			"role": schema.StringAttribute{
				Optional:            true,
				Description:         "Filter members by role.",
				MarkdownDescription: "Filter members by role. One of: `member`, `owner`.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				Description:         "Filter members by status.",
				MarkdownDescription: "Filter members by status.",
			},
			"members": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of members matching the filters.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the member.",
						},
						"email": schema.StringAttribute{
							Computed:    true,
							Description: "The email address of the member.",
						},
						"role": schema.StringAttribute{
							Computed:            true,
							Description:         "The role of the member.",
							MarkdownDescription: "The role of the member. One of: `member`, `owner`.",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "The current status of the member.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the member was created.",
						},
						"updated_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the member was last updated.",
						},
					},
				},
			},
		},
	}
}

func (d *MembersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MembersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	members, err := d.client.ListMembers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing members", err.Error())
		return
	}

	// Apply optional filters
	filterRole := ""
	if !config.Role.IsNull() && !config.Role.IsUnknown() {
		filterRole = config.Role.ValueString()
	}
	filterStatus := ""
	if !config.Status.IsNull() && !config.Status.IsUnknown() {
		filterStatus = config.Status.ValueString()
	}

	var results []MemberDataSourceModel
	for _, m := range members {
		if filterRole != "" && m.Role != filterRole {
			continue
		}
		if filterStatus != "" && m.Status != filterStatus {
			continue
		}

		results = append(results, MemberDataSourceModel{
			ID:        types.StringValue(m.ID),
			Email:     types.StringValue(m.Email),
			Role:      types.StringValue(m.Role),
			Status:    types.StringValue(m.Status),
			CreatedAt: types.StringValue(m.CreatedAt.String()),
			UpdatedAt: types.StringValue(m.UpdatedAt.String()),
		})
	}

	if results == nil {
		results = []MemberDataSourceModel{}
	}

	state := MembersDataSourceModel{
		Role:    config.Role,
		Status:  config.Status,
		Members: results,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
