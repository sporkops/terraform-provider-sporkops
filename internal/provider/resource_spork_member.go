package provider

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sporkops/spork-go"
)

var (
	_ resource.Resource                = &MemberResource{}
	_ resource.ResourceWithConfigure   = &MemberResource{}
	_ resource.ResourceWithImportState = &MemberResource{}
)

type MemberResource struct {
	client *spork.Client
}

type MemberResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func NewMemberResource() resource.Resource {
	return &MemberResource{}
}

func (r *MemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_member"
}

func (r *MemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Spork organization member invitation.",
		MarkdownDescription: "Manages a [Spork](https://sporkops.com) organization **member invitation**.\n\n" +
			"## Lifecycle\n\n" +
			"This resource manages the invitation lifecycle, not a confirmed-member record:\n\n" +
			"  * **Create** sends an invitation email to the supplied address. The resource's `status` starts as `pending` and " +
			"moves to `accepted` once the invitee clicks the emailed link and signs in.\n" +
			"  * **Read** returns the invitation record whether it is pending or accepted. Pending invitations auto-expire " +
			"after 7 days; once expired, the server returns 404 and Terraform removes the resource from state (treating it " +
			"as a destroy-required replacement on the next apply).\n" +
			"  * **Delete** revokes a pending invitation, or removes an accepted member from the organization.\n\n" +
			"The resource name is `spork_member` rather than `spork_member_invitation` because the invitation is the only " +
			"managed unit today — the API does not offer a separate confirmed-member CRUD surface. A future data source " +
			"could expose confirmed members as read-only records if listing becomes useful.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier of the member.",
				MarkdownDescription: "The unique identifier of the member.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Required:            true,
				Description:         "The email address of the member to invite.",
				MarkdownDescription: "The email address of the member to invite.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^.+@.+\..+$`),
						"must be a valid email address",
					),
				},
			},
			"role": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("member"),
				Description:         "The role of the member. Default: member.",
				MarkdownDescription: "The role of the member. One of: `member`, `owner`. Default: `member`.",
				Validators: []validator.String{
					stringvalidator.OneOf("member", "owner"),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				Description:         "The current status of the member.",
				MarkdownDescription: "The current status of the member.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "Timestamp when the member was created.",
				MarkdownDescription: "Timestamp when the member was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				Description:         "Timestamp when the member was last updated.",
				MarkdownDescription: "Timestamp when the member was last updated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *MemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*spork.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *spork.Client, got something else. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = client
}

func (r *MemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.InviteMember(ctx, &spork.InviteMemberInput{
		Email: plan.Email.ValueString(),
		Role:  plan.Role.ValueString(),
	})
	if err != nil {
		addAPIError(&resp.Diagnostics, "Error inviting member", err)
		return
	}

	state := MemberResourceModel{
		ID:        types.StringValue(result.ID),
		Email:     types.StringValue(result.Email),
		Role:      types.StringValue(result.Role),
		Status:    types.StringValue(result.Status),
		CreatedAt: types.StringValue(result.CreatedAt.String()),
		UpdatedAt: types.StringValue(result.UpdatedAt.String()),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	members, err := r.client.ListMembers(ctx)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Error listing members", err)
		return
	}

	var found *spork.Member
	for _, m := range members {
		if m.ID == state.ID.ValueString() {
			found = &m
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState := MemberResourceModel{
		ID:        types.StringValue(found.ID),
		Email:     types.StringValue(found.Email),
		Role:      types.StringValue(found.Role),
		Status:    types.StringValue(found.Status),
		CreatedAt: types.StringValue(found.CreatedAt.String()),
		UpdatedAt: types.StringValue(found.UpdatedAt.String()),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *MemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update supported. Email is ForceNew, role is set on create only.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"The spork_member resource does not support updates. "+
			"Changes to email will trigger a replacement.",
	)
}

func (r *MemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveMember(ctx, state.ID.ValueString())
	if err != nil && !spork.IsNotFound(err) {
		addAPIError(&resp.Diagnostics, "Error removing member", err)
	}
}

func (r *MemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
