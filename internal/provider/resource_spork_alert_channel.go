package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &AlertChannelResource{}
	_ resource.ResourceWithConfigure   = &AlertChannelResource{}
	_ resource.ResourceWithImportState = &AlertChannelResource{}
)

type AlertChannelResource struct {
	client *SporkClient
}

type AlertChannelResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Target   types.String `tfsdk:"target"`
	BotToken types.String `tfsdk:"bot_token"`
	ChatID   types.String `tfsdk:"chat_id"`
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
			"target": schema.StringAttribute{
				Required:            true,
				Description:         "The primary destination for the alert channel. For email: the email address. For webhook/slack/discord/teams/googlechat: the webhook URL. For pagerduty: the integration key. Not used for telegram (use bot_token and chat_id instead).",
				MarkdownDescription: "The primary destination for the alert channel. For `email`: the email address. For `webhook`, `slack`, `discord`, `teams`, `googlechat`: the webhook URL. For `pagerduty`: the integration key. Ignored for `telegram` — use `bot_token` and `chat_id` instead.",
			},
			"bot_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				Description:         "Telegram bot token. Required when type is telegram.",
				MarkdownDescription: "Telegram bot token. Required when `type` is `telegram`.",
			},
			"chat_id": schema.StringAttribute{
				Optional:            true,
				Description:         "Telegram chat ID. Required when type is telegram.",
				MarkdownDescription: "Telegram chat ID. Required when `type` is `telegram`.",
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

	client, ok := req.ProviderData.(*SporkClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *SporkClient, got something else. Please report this issue to the provider developers.",
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

	result, err := r.client.CreateAlertChannel(ctx, alertChannelFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating alert channel", err.Error())
		return
	}

	// On create the API returns the full response (including webhook secret).
	// Pass plan as fallback so user-provided fields are preserved in state.
	state := alertChannelToModel(*result, &plan)
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
		if errors.Is(err, ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading alert channel", err.Error())
		return
	}

	// Pass current state so sensitive fields the API redacts are preserved.
	newState := alertChannelToModel(*result, &state)
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

	result, err := r.client.UpdateAlertChannel(ctx, state.ID.ValueString(), alertChannelFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating alert channel", err.Error())
		return
	}

	// User-provided fields come from plan; secret is preserved from state (API
	// redacts it after creation); verified is refreshed from the API response.
	newState := AlertChannelResourceModel{
		ID:       state.ID,
		Name:     plan.Name,
		Type:     plan.Type,
		Target:   plan.Target,
		BotToken: plan.BotToken,
		ChatID:   plan.ChatID,
		Secret:   state.Secret,
		Verified: types.BoolValue(result.Verified),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *AlertChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AlertChannelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAlertChannel(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting alert channel", err.Error())
	}
}

func (r *AlertChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Conversion helpers

// alertChannelFromModel serializes a Terraform model to the API request struct.
func alertChannelFromModel(model AlertChannelResourceModel) AlertChannel {
	channel := AlertChannel{
		Name:   model.Name.ValueString(),
		Type:   model.Type.ValueString(),
		Config: make(map[string]string),
	}

	switch channel.Type {
	case "email":
		if v := model.Target.ValueString(); v != "" {
			channel.Config["to"] = v
		}
	case "webhook", "slack", "discord", "teams", "googlechat":
		if v := model.Target.ValueString(); v != "" {
			channel.Config["url"] = v
		}
	case "pagerduty":
		if v := model.Target.ValueString(); v != "" {
			channel.Config["integration_key"] = v
		}
	case "telegram":
		if v := model.BotToken.ValueString(); v != "" {
			channel.Config["bot_token"] = v
		}
		if v := model.ChatID.ValueString(); v != "" {
			channel.Config["chat_id"] = v
		}
	}

	return channel
}

// alertChannelToModel deserializes an API AlertChannel into a Terraform model.
// fallback provides current state/plan values for sensitive fields that the API
// redacts on reads (bot_token, integration_key, url, secret).
func alertChannelToModel(c AlertChannel, fallback *AlertChannelResourceModel) AlertChannelResourceModel {
	model := AlertChannelResourceModel{
		ID:       types.StringValue(c.ID),
		Name:     types.StringValue(c.Name),
		Type:     types.StringValue(c.Type),
		Verified: types.BoolValue(c.Verified),
		Target:   types.StringNull(),
		BotToken: types.StringNull(),
		ChatID:   types.StringNull(),
		Secret:   types.StringNull(),
	}

	switch c.Type {
	case "email":
		// Email address is not redacted.
		model.Target = types.StringValue(c.Config["to"])

	case "webhook":
		// URL is masked and secret is deleted on reads; preserve from fallback.
		if v := c.Config["url"]; v != "" {
			model.Target = types.StringValue(v)
		} else if fallback != nil {
			model.Target = fallback.Target
		}
		if v := c.Config["secret"]; v != "" {
			model.Secret = types.StringValue(v)
		} else if fallback != nil {
			model.Secret = fallback.Secret
		}

	case "slack", "discord", "teams", "googlechat":
		// URL is masked on reads; preserve from fallback.
		if v := c.Config["url"]; v != "" {
			model.Target = types.StringValue(v)
		} else if fallback != nil {
			model.Target = fallback.Target
		}

	case "pagerduty":
		// integration_key is deleted on reads; preserve from fallback.
		if v := c.Config["integration_key"]; v != "" {
			model.Target = types.StringValue(v)
		} else if fallback != nil {
			model.Target = fallback.Target
		}

	case "telegram":
		// target is not used for telegram; preserve user-supplied value from fallback.
		if fallback != nil {
			model.Target = fallback.Target
		}
		// bot_token is deleted on reads; preserve from fallback.
		if v := c.Config["bot_token"]; v != "" {
			model.BotToken = types.StringValue(v)
		} else if fallback != nil {
			model.BotToken = fallback.BotToken
		}
		// chat_id is not redacted.
		if v := c.Config["chat_id"]; v != "" {
			model.ChatID = types.StringValue(v)
		} else if fallback != nil {
			model.ChatID = fallback.ChatID
		}
	}

	return model
}
