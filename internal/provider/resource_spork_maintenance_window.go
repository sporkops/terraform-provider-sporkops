package provider

import (
	"context"
	"fmt"
	"time"

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
	_ resource.Resource                   = &MaintenanceWindowResource{}
	_ resource.ResourceWithConfigure      = &MaintenanceWindowResource{}
	_ resource.ResourceWithImportState    = &MaintenanceWindowResource{}
	_ resource.ResourceWithValidateConfig = &MaintenanceWindowResource{}
)

// MaintenanceWindowResource manages `spork_maintenance_window` resources.
//
// Cancel behavior: the `cancelled` attribute is a writeable bool. Flipping
// it from false to true calls CancelMaintenanceWindow server-side and
// moves the window to the "cancelled" state (checks/alerts immediately
// resume for the targeted monitors). The transition is one-way — once
// cancelled, a window stays cancelled; to reinstate, destroy and recreate.
// Matches the pattern used by `spork_monitor.paused`.
type MaintenanceWindowResource struct {
	client *spork.Client
}

type MaintenanceWindowResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	MonitorIDs        types.Set    `tfsdk:"monitor_ids"`
	TagSelectors      types.Set    `tfsdk:"tag_selectors"`
	AllMonitors       types.Bool   `tfsdk:"all_monitors"`
	Timezone          types.String `tfsdk:"timezone"`
	StartAt           types.String `tfsdk:"start_at"`
	EndAt             types.String `tfsdk:"end_at"`
	RecurrenceType    types.String `tfsdk:"recurrence_type"`
	RecurrenceDays    types.List   `tfsdk:"recurrence_days"`
	RecurrenceUntil   types.String `tfsdk:"recurrence_until"`
	SuppressAlerts    types.Bool   `tfsdk:"suppress_alerts"`
	ExcludeFromUptime types.Bool   `tfsdk:"exclude_from_uptime"`
	PauseChecks       types.Bool   `tfsdk:"pause_checks"`
	Cancelled         types.Bool   `tfsdk:"cancelled"`
	State             types.String `tfsdk:"state"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

func NewMaintenanceWindowResource() resource.Resource {
	return &MaintenanceWindowResource{}
}

func (r *MaintenanceWindowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_window"
}

func (r *MaintenanceWindowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a Spork maintenance window that suppresses alerts (and optionally pauses checks) for a set of monitors during a scheduled period.",
		MarkdownDescription: "Manages a [Spork](https://sporkops.com) maintenance window that suppresses alerts (and optionally pauses checks) for a set of monitors during a scheduled period.\n\nTarget **exactly one** of `monitor_ids`, `tag_selectors`, or `all_monitors`. Tag selection uses OR semantics (any matching tag suppresses the monitor).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the maintenance window.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A friendly name for the window.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional description displayed alongside the window.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"monitor_ids": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Monitor IDs the window applies to. Mutually exclusive with tag_selectors and all_monitors.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifierUseStateForUnknown(),
				},
			},
			"tag_selectors": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Monitor tags to match (OR semantics). Mutually exclusive with monitor_ids and all_monitors.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifierUseStateForUnknown(),
				},
			},
			"all_monitors": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Apply the window to every monitor in the organization. Mutually exclusive with monitor_ids and tag_selectors.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"timezone": schema.StringAttribute{
				Required:            true,
				Description:         "IANA timezone name used for display and DST-aware recurrence expansion (e.g., America/Los_Angeles).",
				MarkdownDescription: "IANA timezone name used for display and DST-aware recurrence expansion (e.g., `America/Los_Angeles`).",
			},
			"start_at": schema.StringAttribute{
				Required:    true,
				Description: "Window start (RFC3339 UTC).",
			},
			"end_at": schema.StringAttribute{
				Required:    true,
				Description: "Window end (RFC3339 UTC).",
			},
			"recurrence_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Recurrence type: daily, weekly, or monthly. Leave unset for a one-time window.",
				MarkdownDescription: "Recurrence type: `daily`, `weekly`, or `monthly`. Leave unset for a one-time window.",
				Validators: []validator.String{
					stringvalidator.OneOf("", "daily", "weekly", "monthly"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"recurrence_days": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.Int64Type,
				Description:         "For weekly: days 0-6 (Sun=0). For monthly: days 1-31 (days > last-of-month clamp to the last day).",
				MarkdownDescription: "For `weekly`: days 0-6 (Sun=0). For `monthly`: days 1-31 (days > last-of-month clamp to the last day).",
				PlanModifiers: []planmodifier.List{
					listplanmodifierUseStateForUnknown(),
				},
			},
			"recurrence_until": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Caps recurring series at this RFC3339 UTC timestamp (inclusive).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"suppress_alerts": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Suppress alert deliveries during the window. Defaults to true.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"exclude_from_uptime": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Drop in-window checks from uptime-percentage calculations. Defaults to true.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"pause_checks": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Skip dispatch entirely during the window. Defaults to false (checks keep running so data is preserved).",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"cancelled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Cancel the window. Flip from false to true to cancel. One-way transition — destroy and recreate to reinstate.",
				MarkdownDescription: "Cancel the window. Flip from `false` to `true` to cancel. **One-way transition** — destroy and recreate to reinstate.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "Computed lifecycle state: scheduled, in_progress, completed, or cancelled.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *MaintenanceWindowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MaintenanceWindowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MaintenanceWindowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiW := maintenanceWindowFromModel(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CreateMaintenanceWindow(ctx, &apiW)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Error creating maintenance window", err)
		return
	}

	// A user-set `cancelled = true` at create time must be honored:
	// create first, then cancel in a single apply.
	if !plan.Cancelled.IsNull() && !plan.Cancelled.IsUnknown() && plan.Cancelled.ValueBool() {
		cancelled, err := r.client.CancelMaintenanceWindow(ctx, result.ID)
		if err != nil {
			addAPIError(&resp.Diagnostics, "Error cancelling maintenance window after create", err)
			return
		}
		result = cancelled
	}

	state := maintenanceWindowToModel(ctx, *result, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MaintenanceWindowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MaintenanceWindowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetMaintenanceWindow(ctx, state.ID.ValueString())
	if err != nil {
		if spork.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addAPIError(&resp.Diagnostics, "Error reading maintenance window", err)
		return
	}

	newState := maintenanceWindowToModel(ctx, *result, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *MaintenanceWindowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MaintenanceWindowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MaintenanceWindowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiW := maintenanceWindowFromModel(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateMaintenanceWindow(ctx, state.ID.ValueString(), &apiW)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Error updating maintenance window", err)
		return
	}

	// Handle cancel transition via the dedicated endpoint.
	stateCancelled := !state.Cancelled.IsNull() && state.Cancelled.ValueBool()
	planCancelled := !plan.Cancelled.IsNull() && !plan.Cancelled.IsUnknown() && plan.Cancelled.ValueBool()
	if planCancelled && !stateCancelled {
		cancelled, err := r.client.CancelMaintenanceWindow(ctx, state.ID.ValueString())
		if err != nil {
			addAPIError(&resp.Diagnostics, "Error cancelling maintenance window", err)
			return
		}
		result = cancelled
	}

	newState := maintenanceWindowToModel(ctx, *result, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *MaintenanceWindowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MaintenanceWindowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMaintenanceWindow(ctx, state.ID.ValueString())
	if err != nil && !spork.IsNotFound(err) {
		addAPIError(&resp.Diagnostics, "Error deleting maintenance window", err)
	}
}

// ModifyPlan blocks the cancelled = true → false transition. Cancellation
// is one-way at the server, so Terraform cannot realistically un-cancel —
// without this guard the plan would show a perpetual diff because Read
// always reports `cancelled = true` for cancelled windows.
//
// The guard lives in ModifyPlan (not ValidateConfig) because we need
// access to prior state to compare.
func (r *MaintenanceWindowResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Nothing to check on create (no prior state) or destroy (no plan).
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var state, plan MaintenanceWindowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	stateCancelled := !state.Cancelled.IsNull() && state.Cancelled.ValueBool()
	planCancelled := !plan.Cancelled.IsNull() && !plan.Cancelled.IsUnknown() && plan.Cancelled.ValueBool()
	if stateCancelled && !planCancelled {
		resp.Diagnostics.AddAttributeError(
			path.Root("cancelled"),
			"Cannot un-cancel a maintenance window",
			"Cancellation is a one-way transition. To reinstate a cancelled window, destroy this resource and create a new one (e.g. `terraform taint spork_maintenance_window.example`).",
		)
	}
}

func (r *MaintenanceWindowResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config MaintenanceWindowResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Exactly-one targeting. We can't validate this fully when any of the
	// three attributes is Unknown (computed at apply time), so in that case
	// we defer to the server. When all three are known, reject 0 or >1
	// selections with a clear message.
	targetsKnown := !config.MonitorIDs.IsUnknown() && !config.TagSelectors.IsUnknown() && !config.AllMonitors.IsUnknown()
	if targetsKnown {
		count := 0
		if !config.MonitorIDs.IsNull() && len(config.MonitorIDs.Elements()) > 0 {
			count++
		}
		if !config.TagSelectors.IsNull() && len(config.TagSelectors.Elements()) > 0 {
			count++
		}
		if !config.AllMonitors.IsNull() && config.AllMonitors.ValueBool() {
			count++
		}
		if count != 1 {
			resp.Diagnostics.AddError(
				"Invalid maintenance window targeting",
				"Set exactly one of monitor_ids, tag_selectors, or all_monitors = true.",
			)
		}
	}

	// IANA timezone sanity-check when a literal is supplied.
	if !config.Timezone.IsUnknown() && !config.Timezone.IsNull() {
		tz := config.Timezone.ValueString()
		if tz != "" {
			if _, err := time.LoadLocation(tz); err != nil {
				resp.Diagnostics.AddAttributeError(
					path.Root("timezone"),
					"Invalid IANA timezone",
					fmt.Sprintf("timezone %q is not a valid IANA name: %s", tz, err),
				)
			}
		}
	}

	// RFC3339 sanity on start/end literals.
	for _, pair := range []struct {
		field string
		value types.String
	}{
		{"start_at", config.StartAt},
		{"end_at", config.EndAt},
		{"recurrence_until", config.RecurrenceUntil},
	} {
		if pair.value.IsUnknown() || pair.value.IsNull() {
			continue
		}
		v := pair.value.ValueString()
		if v == "" {
			continue
		}
		if _, err := time.Parse(time.RFC3339, v); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root(pair.field),
				"Invalid RFC3339 timestamp",
				fmt.Sprintf("%s must be an RFC3339 timestamp (e.g., 2026-05-05T09:00:00Z): %s", pair.field, err),
			)
		}
	}
}

func (r *MaintenanceWindowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// maintenanceWindowFromModel serializes a Terraform model to the API request struct.
func maintenanceWindowFromModel(ctx context.Context, model MaintenanceWindowResourceModel, diags *diag.Diagnostics) spork.MaintenanceWindow {
	mw := spork.MaintenanceWindow{
		Name:            model.Name.ValueString(),
		Description:     model.Description.ValueString(),
		Timezone:        model.Timezone.ValueString(),
		StartAt:         model.StartAt.ValueString(),
		EndAt:           model.EndAt.ValueString(),
		RecurrenceType:  model.RecurrenceType.ValueString(),
		RecurrenceUntil: model.RecurrenceUntil.ValueString(),
	}

	if !model.MonitorIDs.IsNull() && !model.MonitorIDs.IsUnknown() {
		var ids []string
		diags.Append(model.MonitorIDs.ElementsAs(ctx, &ids, false)...)
		mw.MonitorIDs = ids
	}
	if !model.TagSelectors.IsNull() && !model.TagSelectors.IsUnknown() {
		var tags []string
		diags.Append(model.TagSelectors.ElementsAs(ctx, &tags, false)...)
		mw.TagSelectors = tags
	}
	if !model.AllMonitors.IsNull() && !model.AllMonitors.IsUnknown() {
		v := model.AllMonitors.ValueBool()
		mw.AllMonitors = &v
	}
	if !model.RecurrenceDays.IsNull() && !model.RecurrenceDays.IsUnknown() {
		var days []int64
		diags.Append(model.RecurrenceDays.ElementsAs(ctx, &days, false)...)
		intDays := make([]int, len(days))
		for i, d := range days {
			intDays[i] = int(d)
		}
		mw.RecurrenceDays = intDays
	}
	if !model.SuppressAlerts.IsNull() && !model.SuppressAlerts.IsUnknown() {
		v := model.SuppressAlerts.ValueBool()
		mw.SuppressAlerts = &v
	}
	if !model.ExcludeFromUptime.IsNull() && !model.ExcludeFromUptime.IsUnknown() {
		v := model.ExcludeFromUptime.ValueBool()
		mw.ExcludeFromUptime = &v
	}
	if !model.PauseChecks.IsNull() && !model.PauseChecks.IsUnknown() {
		v := model.PauseChecks.ValueBool()
		mw.PauseChecks = &v
	}
	return mw
}

// maintenanceWindowToModel deserializes an API MaintenanceWindow into a
// Terraform model. `fallback` preserves user-supplied values for fields
// the API may omit on Read (e.g. empty arrays that the server encodes as
// null) so `terraform plan` after `apply` doesn't produce spurious diffs.
func maintenanceWindowToModel(ctx context.Context, w spork.MaintenanceWindow, fallback *MaintenanceWindowResourceModel, diags *diag.Diagnostics) MaintenanceWindowResourceModel {
	model := MaintenanceWindowResourceModel{
		ID:              types.StringValue(w.ID),
		Name:            types.StringValue(w.Name),
		Description:     types.StringValue(w.Description),
		Timezone:        types.StringValue(w.Timezone),
		StartAt:         types.StringValue(w.StartAt),
		EndAt:           types.StringValue(w.EndAt),
		RecurrenceType:  types.StringValue(w.RecurrenceType),
		RecurrenceUntil: types.StringValue(w.RecurrenceUntil),
		State:           types.StringValue(w.State),
		CreatedAt:       types.StringValue(w.CreatedAt),
		UpdatedAt:       types.StringValue(w.UpdatedAt),
	}

	// MonitorIDs — fall back to prior state when the API omits it (null set)
	// so a user who configured ["m1"] doesn't see perpetual drift if the
	// server encodes the empty case as null on Read.
	if len(w.MonitorIDs) > 0 {
		monIDs, d := types.SetValueFrom(ctx, types.StringType, w.MonitorIDs)
		diags.Append(d...)
		model.MonitorIDs = monIDs
	} else if fallback != nil && !fallback.MonitorIDs.IsNull() && !fallback.MonitorIDs.IsUnknown() {
		model.MonitorIDs = fallback.MonitorIDs
	} else {
		model.MonitorIDs = types.SetNull(types.StringType)
	}

	// TagSelectors — same fallback pattern.
	if len(w.TagSelectors) > 0 {
		tags, d := types.SetValueFrom(ctx, types.StringType, w.TagSelectors)
		diags.Append(d...)
		model.TagSelectors = tags
	} else if fallback != nil && !fallback.TagSelectors.IsNull() && !fallback.TagSelectors.IsUnknown() {
		model.TagSelectors = fallback.TagSelectors
	} else {
		model.TagSelectors = types.SetNull(types.StringType)
	}

	// RecurrenceDays — same fallback pattern (already present before, but
	// rewritten for consistency with the other two).
	if len(w.RecurrenceDays) > 0 {
		days := make([]int64, len(w.RecurrenceDays))
		for i, d := range w.RecurrenceDays {
			days[i] = int64(d)
		}
		list, d := types.ListValueFrom(ctx, types.Int64Type, days)
		diags.Append(d...)
		model.RecurrenceDays = list
	} else if fallback != nil && !fallback.RecurrenceDays.IsNull() && !fallback.RecurrenceDays.IsUnknown() {
		model.RecurrenceDays = fallback.RecurrenceDays
	} else {
		model.RecurrenceDays = types.ListNull(types.Int64Type)
	}

	// Pointer bools — fall back to state when server omits them
	if w.AllMonitors != nil {
		model.AllMonitors = types.BoolValue(*w.AllMonitors)
	} else if fallback != nil && !fallback.AllMonitors.IsNull() && !fallback.AllMonitors.IsUnknown() {
		model.AllMonitors = fallback.AllMonitors
	} else {
		model.AllMonitors = types.BoolValue(false)
	}

	if w.SuppressAlerts != nil {
		model.SuppressAlerts = types.BoolValue(*w.SuppressAlerts)
	} else if fallback != nil && !fallback.SuppressAlerts.IsNull() && !fallback.SuppressAlerts.IsUnknown() {
		model.SuppressAlerts = fallback.SuppressAlerts
	} else {
		model.SuppressAlerts = types.BoolValue(true)
	}

	if w.ExcludeFromUptime != nil {
		model.ExcludeFromUptime = types.BoolValue(*w.ExcludeFromUptime)
	} else if fallback != nil && !fallback.ExcludeFromUptime.IsNull() && !fallback.ExcludeFromUptime.IsUnknown() {
		model.ExcludeFromUptime = fallback.ExcludeFromUptime
	} else {
		model.ExcludeFromUptime = types.BoolValue(true)
	}

	if w.PauseChecks != nil {
		model.PauseChecks = types.BoolValue(*w.PauseChecks)
	} else if fallback != nil && !fallback.PauseChecks.IsNull() && !fallback.PauseChecks.IsUnknown() {
		model.PauseChecks = fallback.PauseChecks
	} else {
		model.PauseChecks = types.BoolValue(false)
	}

	// Cancelled is derived from state
	model.Cancelled = types.BoolValue(w.State == spork.MaintenanceStateCancelled)

	return model
}
