package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sporkops/spork-go"
)

var (
	_ resource.Resource                   = &AlertChannelResource{}
	_ resource.ResourceWithConfigure      = &AlertChannelResource{}
	_ resource.ResourceWithImportState    = &AlertChannelResource{}
	_ resource.ResourceWithValidateConfig = &AlertChannelResource{}
)

type AlertChannelResource struct {
	client *spork.Client
}

type AlertChannelResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Config   types.Map    `tfsdk:"config"`
	Secret   types.String `tfsdk:"secret"`
	Verified types.Bool   `tfsdk:"verified"`
}

func NewAlertChannelResource() resource.Resource {
	return &AlertChannelResource{}
}

func (r *AlertChannelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_channel"
}

func (r *AlertChannelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a Spork alert channel for uptime notifications.",
		MarkdownDescription: "Manages a [Spork](https://sporkops.com) alert channel that receives notifications when a monitor detects downtime.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier of the alert channel.",
				MarkdownDescription: "The unique identifier of the alert channel.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A friendly name for the alert channel.",
			},
			"type": schema.StringAttribute{
				Required:            true,
				Description:         "The channel type. Changing this forces a new resource.",
				MarkdownDescription: "The channel type: `email`, `webhook`, `slack`, `discord`, `teams`, `pagerduty`, `telegram`, or `googlechat`. Changing this forces a new resource.",
				Validators: []validator.String{
					stringvalidator.OneOf("email", "webhook", "slack", "discord", "teams", "pagerduty", "telegram", "googlechat"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config": schema.MapAttribute{
				Required:            true,
				Sensitive:           true,
				ElementType:         types.StringType,
				Description:         "Channel configuration as key-value pairs. Keys depend on channel type: email: {to}, webhook: {url}, slack/discord/teams/googlechat: {url}, pagerduty: {integration_key}, telegram: {bot_token, chat_id}.",
				MarkdownDescription: "Channel configuration as key-value pairs. Keys depend on channel type: `email`: `{to}`, `webhook`: `{url}`, `slack`/`discord`/`teams`/`googlechat`: `{url}`, `pagerduty`: `{integration_key}`, `telegram`: `{bot_token, chat_id}`.",
			},
			"secret": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				Description:         "Webhook signing secret. Only returned for webhook type, shown once at creation.",
				MarkdownDescription: "Webhook signing secret. Only returned for `webhook` type, shown once at creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"verified": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the channel has been verified. Relevant for email type.",
				MarkdownDescription: "Whether the channel has been verified. Relevant for `email` type.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *AlertChannelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AlertChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AlertChannelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiChannel := alertChannelFromModel(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.client.CreateAlertChannel(ctx, &apiChannel)
	if err != nil {
		resp.Diagnostics.AddError("Error creating alert channel", err.Error())
		return
	}

	// On create the API returns the full response (including webhook secret).
	// Pass plan as fallback so user-provided fields are preserved in state.
	state := alertChannelToModel(ctx, *result, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AlertChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AlertChannelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetAlertChannel(ctx, state.ID.ValueString())
	if err != nil {
		if spork.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading alert channel", err.Error())
		return
	}

	// Pass current state so sensitive fields the API redacts are preserved.
	newState := alertChannelToModel(ctx, *result, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *AlertChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AlertChannelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AlertChannelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiChannel := alertChannelFromModel(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.client.UpdateAlertChannel(ctx, state.ID.ValueString(), &apiChannel)
	if err != nil {
		resp.Diagnostics.AddError("Error updating alert channel", err.Error())
		return
	}

	// Use plan as fallback for redacted config values; preserve secret from state.
	newState := alertChannelToModel(ctx, *result, &plan, &resp.Diagnostics)
	newState.Secret = state.Secret
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *AlertChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AlertChannelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAlertChannel(ctx, state.ID.ValueString())
	if err != nil && !spork.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting alert channel", err.Error())
	}
}

func (r *AlertChannelResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config AlertChannelResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip validation if type or config is unknown (e.g., from a variable)
	if config.Type.IsUnknown() || config.Config.IsUnknown() || config.Config.IsNull() {
		return
	}

	channelType := config.Type.ValueString()
	var configKeys map[string]string
	resp.Diagnostics.Append(config.Config.ElementsAs(ctx, &configKeys, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required config keys by channel type
	var requiredKeys []string
	switch channelType {
	case "email":
		requiredKeys = []string{"to"}
	case "webhook", "slack", "discord", "teams", "googlechat":
		requiredKeys = []string{"url"}
	case "pagerduty":
		requiredKeys = []string{"integration_key"}
	case "telegram":
		requiredKeys = []string{"bot_token", "chat_id"}
	}

	for _, key := range requiredKeys {
		if val, ok := configKeys[key]; !ok || val == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("config"),
				"Missing Required Config Key",
				fmt.Sprintf("config key %q is required for channel type %q", key, channelType),
			)
		}
	}
}

func (r *AlertChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Conversion helpers

// alertChannelFromModel serializes a Terraform model to the API request struct.
func alertChannelFromModel(ctx context.Context, model AlertChannelResourceModel, diags *diag.Diagnostics) spork.AlertChannel {
	channel := spork.AlertChannel{
		Name:   model.Name.ValueString(),
		Type:   model.Type.ValueString(),
		Config: make(map[string]string),
	}

	if !model.Config.IsNull() && !model.Config.IsUnknown() {
		diags.Append(model.Config.ElementsAs(ctx, &channel.Config, false)...)
	}

	return channel
}

// alertChannelToModel deserializes an API AlertChannel into a Terraform model.
// fallback provides current state/plan values for sensitive fields that the API
// redacts on reads (bot_token, integration_key, url, etc.).
func alertChannelToModel(ctx context.Context, c spork.AlertChannel, fallback *AlertChannelResourceModel, diags *diag.Diagnostics) AlertChannelResourceModel {
	model := AlertChannelResourceModel{
		ID:       types.StringValue(c.ID),
		Name:     types.StringValue(c.Name),
		Type:     types.StringValue(c.Type),
		Verified: types.BoolValue(c.Verified),
		Secret:   types.StringNull(),
	}

	// Build config from API response, extracting secret separately.
	configData := make(map[string]string)
	for k, v := range c.Config {
		if k == "secret" {
			model.Secret = types.StringValue(v)
		} else {
			configData[k] = v
		}
	}

	// Secret not in API response — preserve from state.
	if model.Secret.IsNull() && fallback != nil {
		model.Secret = fallback.Secret
	}

	// Non-webhook channels never have a secret; ensure the field is set to avoid
	// Terraform erroring with "Provider returned invalid result object after apply."
	if model.Secret.IsNull() || model.Secret.IsUnknown() {
		model.Secret = types.StringValue("")
	}

	// Preserve state/plan values for config keys the API redacted on GET.
	if fallback != nil && !fallback.Config.IsNull() && !fallback.Config.IsUnknown() {
		var fallbackConfig map[string]string
		fallback.Config.ElementsAs(ctx, &fallbackConfig, false)
		for k, v := range fallbackConfig {
			existing, ok := configData[k]
			if !ok || isRedacted(existing, v) {
				configData[k] = v
			}
		}
	}

	configMap, d := types.MapValueFrom(ctx, types.StringType, configData)
	diags.Append(d...)
	model.Config = configMap

	return model
}

// isRedacted reports whether apiValue looks like a redacted version of
// stateValue. The API masks sensitive config fields on GET responses by
// truncating them, inserting "..." placeholders, or stripping them entirely.
func isRedacted(apiValue, stateValue string) bool {
	if apiValue == "" {
		return true
	}
	if len(apiValue) < len(stateValue) {
		return true
	}
	if strings.Contains(apiValue, "...") {
		return true
	}
	return false
}
