package provider

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sporkops/spork-go"
)

var (
	_ resource.Resource                = &StatusPageResource{}
	_ resource.ResourceWithConfigure   = &StatusPageResource{}
	_ resource.ResourceWithImportState = &StatusPageResource{}
)

type StatusPageResource struct {
	client *spork.Client
}

type StatusPageResourceModel struct {
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
	Password                types.String `tfsdk:"password"`
	CreatedAt               types.String `tfsdk:"created_at"`
	UpdatedAt               types.String `tfsdk:"updated_at"`
}

type StatusPageComponentModel struct {
	ID          types.String `tfsdk:"id"`
	MonitorID   types.String `tfsdk:"monitor_id"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	Group       types.String `tfsdk:"group"`
	Order       types.Int64  `tfsdk:"order"`
}

type StatusPageComponentGroupModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Order       types.Int64  `tfsdk:"order"`
}

var componentAttrTypes = map[string]attr.Type{
	"id":           types.StringType,
	"monitor_id":   types.StringType,
	"display_name": types.StringType,
	"description":  types.StringType,
	"group":        types.StringType,
	"order":        types.Int64Type,
}

var componentGroupAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"order":       types.Int64Type,
}

func NewStatusPageResource() resource.Resource {
	return &StatusPageResource{}
}

func (r *StatusPageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

func (r *StatusPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a Spork status page.",
		MarkdownDescription: "Manages a [Spork](https://sporkops.com) public status page with components, custom domains, and branding.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the status page.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the status page (1-100 characters).",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
			},
			"slug": schema.StringAttribute{
				Required:            true,
				Description:         "URL-safe slug for the status page (2-63 lowercase alphanumeric characters or hyphens).",
				MarkdownDescription: "URL-safe slug for the status page. Used in the URL: `https://<slug>.status.sporkops.com`. Must be 2-63 lowercase alphanumeric characters or hyphens.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`),
						"must be 2-63 lowercase alphanumeric characters or hyphens, cannot start or end with a hyphen",
					),
				},
			},
			"custom_domain": schema.StringAttribute{
				Optional:            true,
				Description:         "Custom domain for the status page.",
				MarkdownDescription: "Custom domain for the status page (e.g. `status.example.com`). Requires a CNAME record pointing to `status.sporkops.com`.",
			},
			"domain_status": schema.StringAttribute{
				Computed:            true,
				Description:         "Verification status of the custom domain: pending, active, or failed.",
				MarkdownDescription: "Verification status of the custom domain: `pending`, `active`, or `failed`.",
			},
			"theme": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("light"),
				Description:         "Color theme for the status page. Default: light.",
				MarkdownDescription: "Color theme for the status page. One of: `light`, `dark`, `blue`, `midnight`. Default: `light`.",
				Validators: []validator.String{
					stringvalidator.OneOf("light", "dark", "blue", "midnight"),
				},
			},
			"accent_color": schema.StringAttribute{
				Optional:            true,
				Description:         "Accent color for the status page as a hex color (e.g. #ff0000).",
				MarkdownDescription: "Accent color for the status page as a hex color (e.g. `#ff0000`).",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^#[0-9a-fA-F]{6}$`),
						"must be a hex color (e.g. #ff0000)",
					),
				},
			},
			"font_family": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("system"),
				Description:         "Font family for the status page. Default: system.",
				MarkdownDescription: "Font family for the status page. One of: `system`, `sans-serif`, `serif`, `monospace`. Default: `system`.",
				Validators: []validator.String{
					stringvalidator.OneOf("system", "sans-serif", "serif", "monospace"),
				},
			},
			"header_style": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
				Description:         "Header style for the status page. Default: default.",
				MarkdownDescription: "Header style for the status page. One of: `default`, `banner`, `minimal`. Default: `default`.",
				Validators: []validator.String{
					stringvalidator.OneOf("default", "banner", "minimal"),
				},
			},
			"logo_url": schema.StringAttribute{
				Optional:            true,
				Description:         "URL of the logo to display on the status page. Must be an https:// URL.",
				MarkdownDescription: "URL of the logo to display on the status page. Must be an `https://` URL.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^https://`),
						"must be an https:// URL",
					),
				},
			},
			"webhook_url": schema.StringAttribute{
				Optional:            true,
				Description:         "Webhook URL for incident notifications. Must be an https:// URL.",
				MarkdownDescription: "Webhook URL for incident notifications. Must be an `https://` URL.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^https://`),
						"must be an https:// URL",
					),
				},
			},
			"email_subscribers_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether email subscriber notifications are enabled. Default: false.",
				MarkdownDescription: "Whether email subscriber notifications are enabled on the public status page. Default: `false`.",
			},
			"is_public": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the status page is publicly accessible. Default: true.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password for private status pages. Only used when is_public is false. Visitors must enter this password to view the page.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the status page was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the status page was last updated.",
			},
		},
		Blocks: map[string]schema.Block{
			"components": schema.ListNestedBlock{
				Description:         "Components displayed on the status page. Each component maps a monitor to a display name.",
				MarkdownDescription: "Components displayed on the status page. Each component maps a monitor to a display name.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the component.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"monitor_id": schema.StringAttribute{
							Required:    true,
							Description: "The ID of the monitor to display on the status page.",
						},
						"display_name": schema.StringAttribute{
							Required:    true,
							Description: "The display name shown on the status page for this component.",
						},
						"description": schema.StringAttribute{
							Optional:    true,
							Description: "A description of the component.",
						},
						"group": schema.StringAttribute{
							Optional:            true,
							Description:         "The name of the component group this component belongs to.",
							MarkdownDescription: "The name of the component group this component belongs to. Must match a `name` from a `component_groups` entry.",
						},
						"order": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "Display order of the component on the status page.",
						},
					},
				},
			},
			"component_groups": schema.ListNestedBlock{
				Description:         "Component groups for organizing components into named sections.",
				MarkdownDescription: "Component groups for organizing components into named sections on the status page.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the component group.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The display name of the component group.",
						},
						"description": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "A description of the component group.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"order": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "Display order of the component group on the status page.",
						},
					},
				},
			},
		},
	}
}

func (r *StatusPageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StatusPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StatusPageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPage := statusPageFromModel(ctx, plan)

	result, err := r.client.CreateStatusPage(ctx, &apiPage)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Error creating status page", err)
		return
	}

	// Save state immediately so the resource is tracked even if the
	// custom domain call below fails (prevents orphaned resources).
	state := statusPageToModel(ctx, *result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set custom domain if specified
	if !plan.CustomDomain.IsNull() && !plan.CustomDomain.IsUnknown() && plan.CustomDomain.ValueString() != "" {
		err := r.client.SetCustomDomain(ctx, result.ID, plan.CustomDomain.ValueString())
		if err != nil {
			addAPIError(&resp.Diagnostics, "Error setting custom domain", err)
			return
		}
		// Re-read to get domain_status
		result, err = r.client.GetStatusPage(ctx, result.ID)
		if err != nil {
			addAPIError(&resp.Diagnostics, "Error reading status page after setting custom domain", err)
			return
		}
		state = statusPageToModel(ctx, *result)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (r *StatusPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StatusPageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetStatusPage(ctx, state.ID.ValueString())
	if err != nil {
		if spork.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addAPIError(&resp.Diagnostics, "Error reading status page", err)
		return
	}

	newState := statusPageToModel(ctx, *result)
	// Preserve password from state — the API never returns it
	newState.Password = state.Password
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *StatusPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StatusPageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state StatusPageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPage := statusPageFromModel(ctx, plan)

	result, err := r.client.UpdateStatusPage(ctx, state.ID.ValueString(), &apiPage)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Error updating status page", err)
		return
	}

	// Handle custom domain changes
	oldDomain := state.CustomDomain.ValueString()
	newDomain := plan.CustomDomain.ValueString()
	if plan.CustomDomain.IsNull() {
		newDomain = ""
	}

	if oldDomain != newDomain {
		if oldDomain != "" && newDomain == "" {
			// Remove custom domain
			err := r.client.RemoveCustomDomain(ctx, result.ID)
			if err != nil {
				addAPIError(&resp.Diagnostics, "Error removing custom domain", err)
				return
			}
		} else if newDomain != "" {
			// Set (or change) custom domain
			err := r.client.SetCustomDomain(ctx, result.ID, newDomain)
			if err != nil {
				addAPIError(&resp.Diagnostics, "Error setting custom domain", err)
				return
			}
		}
		// Re-read to get updated domain state
		result, err = r.client.GetStatusPage(ctx, result.ID)
		if err != nil {
			addAPIError(&resp.Diagnostics, "Error reading status page after domain change", err)
			return
		}
	}

	newState := statusPageToModel(ctx, *result)
	newState.Password = plan.Password
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *StatusPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StatusPageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteStatusPage(ctx, state.ID.ValueString())
	if err != nil && !spork.IsNotFound(err) {
		addAPIError(&resp.Diagnostics, "Error deleting status page", err)
	}
}

func (r *StatusPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Conversion helpers

func statusPageFromModel(ctx context.Context, model StatusPageResourceModel) spork.StatusPage {
	page := spork.StatusPage{
		Name:                    model.Name.ValueString(),
		Slug:                    model.Slug.ValueString(),
		Theme:                   model.Theme.ValueString(),
		IsPublic:                model.IsPublic.ValueBool(),
		EmailSubscribersEnabled: model.EmailSubscribersEnabled.ValueBool(),
	}

	if !model.AccentColor.IsNull() && !model.AccentColor.IsUnknown() {
		page.AccentColor = model.AccentColor.ValueString()
	}

	page.FontFamily = model.FontFamily.ValueString()
	page.HeaderStyle = model.HeaderStyle.ValueString()

	if !model.LogoURL.IsNull() && !model.LogoURL.IsUnknown() {
		page.LogoURL = model.LogoURL.ValueString()
	}

	if !model.WebhookURL.IsNull() && !model.WebhookURL.IsUnknown() {
		page.WebhookURL = model.WebhookURL.ValueString()
	}

	if !model.Password.IsNull() && !model.Password.IsUnknown() {
		page.Password = model.Password.ValueString()
	}

	if !model.Components.IsNull() && !model.Components.IsUnknown() {
		var components []StatusPageComponentModel
		model.Components.ElementsAs(ctx, &components, false)
		for _, c := range components {
			comp := spork.StatusComponent{
				MonitorID:   c.MonitorID.ValueString(),
				DisplayName: c.DisplayName.ValueString(),
				Order:       int(c.Order.ValueInt64()),
			}
			if !c.ID.IsNull() && !c.ID.IsUnknown() {
				comp.ID = c.ID.ValueString()
			}
			if !c.Description.IsNull() && !c.Description.IsUnknown() {
				comp.Description = c.Description.ValueString()
			}
			if !c.Group.IsNull() && !c.Group.IsUnknown() {
				comp.GroupName = c.Group.ValueString()
			}
			page.Components = append(page.Components, comp)
		}
	}

	if !model.ComponentGroups.IsNull() && !model.ComponentGroups.IsUnknown() {
		var groups []StatusPageComponentGroupModel
		model.ComponentGroups.ElementsAs(ctx, &groups, false)
		for _, g := range groups {
			group := spork.ComponentGroup{
				Name:  g.Name.ValueString(),
				Order: int(g.Order.ValueInt64()),
			}
			if !g.ID.IsNull() && !g.ID.IsUnknown() {
				group.ID = g.ID.ValueString()
			}
			if !g.Description.IsNull() && !g.Description.IsUnknown() {
				group.Description = g.Description.ValueString()
			}
			page.ComponentGroups = append(page.ComponentGroups, group)
		}
	}

	return page
}

func statusPageToModel(_ context.Context, p spork.StatusPage) StatusPageResourceModel {
	model := StatusPageResourceModel{
		ID:                      types.StringValue(p.ID),
		Name:                    types.StringValue(p.Name),
		Slug:                    types.StringValue(p.Slug),
		Theme:                   types.StringValue(p.Theme),
		IsPublic:                types.BoolValue(p.IsPublic),
		EmailSubscribersEnabled: types.BoolValue(p.EmailSubscribersEnabled),
		CreatedAt:               types.StringValue(p.CreatedAt),
		UpdatedAt:               types.StringValue(p.UpdatedAt),
	}

	// Custom domain
	if p.CustomDomain != "" {
		model.CustomDomain = types.StringValue(p.CustomDomain)
		model.DomainStatus = types.StringValue(p.DomainStatus)
	} else {
		model.CustomDomain = types.StringNull()
		model.DomainStatus = types.StringNull()
	}

	// Font family & header style
	model.FontFamily = types.StringValue(p.FontFamily)
	model.HeaderStyle = types.StringValue(p.HeaderStyle)

	// Accent color
	if p.AccentColor != "" {
		model.AccentColor = types.StringValue(p.AccentColor)
	} else {
		model.AccentColor = types.StringNull()
	}

	// Logo URL
	if p.LogoURL != "" {
		model.LogoURL = types.StringValue(p.LogoURL)
	} else {
		model.LogoURL = types.StringNull()
	}

	// Webhook URL
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
		model.Components = types.ListValueMust(types.ObjectType{AttrTypes: componentAttrTypes}, []attr.Value{})
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
		model.ComponentGroups = types.ListValueMust(types.ObjectType{AttrTypes: componentGroupAttrTypes}, []attr.Value{})
	}

	return model
}
