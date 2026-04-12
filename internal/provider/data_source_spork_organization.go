package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sporkops/spork-go"
)

var _ datasource.DataSource = &OrganizationDataSource{}
var _ datasource.DataSourceWithConfigure = &OrganizationDataSource{}

type OrganizationDataSource struct {
	client *spork.Client
}

type OrganizationDataSourceModel struct {
	ID            types.String                      `tfsdk:"id"`
	Name          types.String                      `tfsdk:"name"`
	CreatedAt     types.String                      `tfsdk:"created_at"`
	UpdatedAt     types.String                      `tfsdk:"updated_at"`
	Subscriptions []SubscriptionDataSourceModel     `tfsdk:"subscriptions"`
}

type SubscriptionDataSourceModel struct {
	Product           types.String `tfsdk:"product"`
	Plan              types.String `tfsdk:"plan"`
	Entitlements      types.Map    `tfsdk:"entitlements"`
	HasPaymentMethod  types.Bool   `tfsdk:"has_payment_method"`
	CancelAtPeriodEnd types.Bool   `tfsdk:"cancel_at_period_end"`
	CancelAt          types.String `tfsdk:"cancel_at"`
	TrialEndsAt       types.String `tfsdk:"trial_ends_at"`
}

func NewOrganizationDataSource() datasource.DataSource {
	return &OrganizationDataSource{}
}

func (d *OrganizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *OrganizationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches the authenticated user's Spork organization.",
		MarkdownDescription: "Fetches the authenticated user's [Spork](https://sporkops.com) organization, including subscription details.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the organization.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the organization.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the organization was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the organization was last updated.",
			},
			"subscriptions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of product subscriptions for the organization.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"product": schema.StringAttribute{
							Computed:    true,
							Description: "The product name.",
						},
						"plan": schema.StringAttribute{
							Computed:    true,
							Description: "The subscription plan.",
						},
						"entitlements": schema.MapAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Entitlements for the subscription as key-value pairs.",
						},
						"has_payment_method": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether a payment method is configured.",
						},
						"cancel_at_period_end": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the subscription will cancel at the end of the billing period.",
						},
						"cancel_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the subscription will be cancelled, if applicable.",
						},
						"trial_ends_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the trial ends, if applicable.",
						},
					},
				},
			},
		},
	}
}

func (d *OrganizationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	org, err := d.client.GetOrganization(ctx)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Error reading organization", err)
		return
	}

	var subscriptions []SubscriptionDataSourceModel
	for _, s := range org.Subscriptions {
		// Convert entitlements map[string]any to map[string]string
		entitlementStrings := make(map[string]string, len(s.Entitlements))
		for k, v := range s.Entitlements {
			entitlementStrings[k] = fmt.Sprintf("%v", v)
		}

		entitlements, diags := types.MapValueFrom(ctx, types.StringType, entitlementStrings)
		resp.Diagnostics.Append(diags...)

		cancelAt := types.StringNull()
		if s.CancelAt != nil {
			cancelAt = types.StringValue(s.CancelAt.String())
		}
		trialEndsAt := types.StringNull()
		if s.TrialEndsAt != nil {
			trialEndsAt = types.StringValue(s.TrialEndsAt.String())
		}

		subscriptions = append(subscriptions, SubscriptionDataSourceModel{
			Product:           types.StringValue(s.Product),
			Plan:              types.StringValue(s.Plan),
			Entitlements:      entitlements,
			HasPaymentMethod:  types.BoolValue(s.HasPaymentMethod),
			CancelAtPeriodEnd: types.BoolValue(s.CancelAtPeriodEnd),
			CancelAt:          cancelAt,
			TrialEndsAt:       trialEndsAt,
		})
	}

	if subscriptions == nil {
		subscriptions = []SubscriptionDataSourceModel{}
	}

	state := OrganizationDataSourceModel{
		ID:            types.StringValue(org.ID),
		Name:          types.StringValue(org.Name),
		CreatedAt:     types.StringValue(org.CreatedAt.String()),
		UpdatedAt:     types.StringValue(org.UpdatedAt.String()),
		Subscriptions: subscriptions,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
