package provider

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sporkops/spork-go"
)

var (
	_ resource.Resource                   = &MonitorResource{}
	_ resource.ResourceWithConfigure      = &MonitorResource{}
	_ resource.ResourceWithImportState    = &MonitorResource{}
	_ resource.ResourceWithValidateConfig = &MonitorResource{}
)

type MonitorResource struct {
	client *spork.Client
}

type MonitorResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Target          types.String `tfsdk:"target"`
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
	Method          types.String `tfsdk:"method"`
	ExpectedStatus  types.Int64  `tfsdk:"expected_status"`
	Interval        types.Int64  `tfsdk:"interval"`
	Timeout         types.Int64  `tfsdk:"timeout"`
	Regions         types.Set    `tfsdk:"regions"`
	AlertChannelIDs types.Set    `tfsdk:"alert_channel_ids"`
	Tags            types.Set    `tfsdk:"tags"`
	Paused          types.Bool   `tfsdk:"paused"`
	Status          types.String `tfsdk:"status"`
	Headers         types.Map    `tfsdk:"headers"`
	Body            types.String `tfsdk:"body"`
	Keyword         types.String `tfsdk:"keyword"`
	KeywordType     types.String `tfsdk:"keyword_type"`
	SSLWarnDays     types.Int64  `tfsdk:"ssl_warn_days"`
}

func NewMonitorResource() resource.Resource {
	return &MonitorResource{}
}

func (r *MonitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (r *MonitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a Spork uptime monitor.",
		MarkdownDescription: "Manages a [Spork](https://sporkops.com) uptime monitor that periodically checks a target URL and reports its status.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier of the monitor.",
				MarkdownDescription: "The unique identifier of the monitor.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"target": schema.StringAttribute{
				Required:            true,
				Description:         "The target to monitor. For http, keyword, and ssl types: a URL starting with http:// or https://. For dns and ping: a bare hostname or IP (e.g., example.com). For tcp: host:port (e.g., example.com:443).",
				MarkdownDescription: "The target to monitor.\n\n  * `http`, `keyword`, `ssl`: URL starting with `http://` or `https://` (e.g., `https://example.com`).\n  * `dns`, `ping`: bare hostname or IP (e.g., `example.com`). No URL scheme, path, or port.\n  * `tcp`: `host:port` (e.g., `example.com:443`).",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A human-readable name for the monitor (1-100 characters).",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("http"),
				Description:         "Monitor type. Default: http.",
				MarkdownDescription: "Monitor type. One of: `http`, `ssl`, `dns`, `keyword`, `tcp`, `ping`. Default: `http`.",
				Validators: []validator.String{
					stringvalidator.OneOf("http", "ssl", "dns", "keyword", "tcp", "ping"),
				},
			},
			"method": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "HTTP method for http and keyword monitors. Server defaults to GET if not specified.",
				MarkdownDescription: "HTTP method for `http` and `keyword` monitors. One of: `GET`, `HEAD`, `POST`, `PUT`. Server defaults to `GET` if not specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("GET", "HEAD", "POST", "PUT"),
				},
			},
			"expected_status": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "Expected HTTP status code for http and keyword monitors. Server defaults to 200 if not specified.",
				MarkdownDescription: "Expected HTTP status code (100-599) for `http` and `keyword` monitors. Server defaults to `200` if not specified.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(100, 599),
				},
			},
			"interval": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
				Description:         "Check interval in seconds (60-86400, must be a multiple of 60). Default: 60.",
				MarkdownDescription: "Check interval in seconds (`60`-`86400`, must be a multiple of 60). Default: `60`.",
				Validators: []validator.Int64{
					int64validator.Between(60, 86400),
					MultipleOf(60),
				},
			},
			"timeout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(30),
				Description:         "Timeout in seconds for each check (5-120). Default: 30.",
				MarkdownDescription: "Timeout in seconds for each check (`5`-`120`). Default: `30`.",
				Validators: []validator.Int64{
					int64validator.Between(5, 120),
				},
			},
			"regions": schema.SetAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Description:         "Regions to check from. Default: [\"us-central1\"]. Order is not significant.",
				MarkdownDescription: "Regions to check from. Available: `us-central1`, `europe-west1`. Default: `[\"us-central1\"]`. Order is not significant.",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf("us-central1", "europe-west1"),
					),
				},
			},
			"alert_channel_ids": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "IDs of alert channels to notify on status changes. Order is not significant.",
			},
			"tags": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags for organizing monitors. Order is not significant.",
			},
			"paused": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description:         "Whether the monitor is paused.",
				MarkdownDescription: "Whether the monitor is paused.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				Description:         "Current monitor status: up, down, degraded, paused, or pending.",
				MarkdownDescription: "Current monitor status: `up`, `down`, `degraded`, `paused`, or `pending`.",
			},
			"headers": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "HTTP headers to send with each check.",
			},
			"body": schema.StringAttribute{
				Optional:    true,
				Description: "Request body to send with POST/PUT checks.",
			},
			"keyword": schema.StringAttribute{
				Optional:    true,
				Description: "Keyword to search for in the response body.",
			},
			"keyword_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the keyword should exist or not exist in the response. Only used for keyword monitors.",
				MarkdownDescription: "Whether the keyword should `exists` or `not_exists` in the response. Only used for `keyword` monitors.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("exists", "not_exists"),
				},
			},
			"ssl_warn_days": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of days before SSL expiry to trigger a warning. Only used for ssl monitors.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *MonitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiMonitor := monitorFromModel(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CreateMonitor(ctx, &apiMonitor)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Error creating monitor", err)
		return
	}

	state := monitorToModel(ctx, *result, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetMonitor(ctx, state.ID.ValueString())
	if err != nil {
		if spork.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addAPIError(&resp.Diagnostics, "Error reading monitor", err)
		return
	}

	newState := monitorToModel(ctx, *result, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *MonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiMonitor := monitorFromModel(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateMonitor(ctx, state.ID.ValueString(), &apiMonitor)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Error updating monitor", err)
		return
	}

	newState := monitorToModel(ctx, *result, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *MonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, state.ID.ValueString())
	if err != nil && !spork.IsNotFound(err) {
		addAPIError(&resp.Diagnostics, "Error deleting monitor", err)
	}
}

func (r *MonitorResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config MonitorResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip validation if type is unknown (e.g., from a variable)
	if config.Type.IsUnknown() {
		return
	}

	monType := config.Type.ValueString()

	// keyword monitors require keyword and keyword_type
	if monType == "keyword" {
		if config.Keyword.IsNull() || config.Keyword.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("keyword"),
				"Missing Required Attribute",
				"keyword is required when type is \"keyword\". "+
					"Set keyword to the text to search for in the response body.",
			)
		}
		if config.KeywordType.IsNull() || config.KeywordType.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("keyword_type"),
				"Missing Required Attribute",
				"keyword_type is required when type is \"keyword\". "+
					"Set keyword_type to \"exists\" or \"not_exists\".",
			)
		}
	}

	// Warn if keyword/keyword_type set on non-keyword monitors
	if monType != "keyword" && monType != "" {
		if !config.Keyword.IsNull() && !config.Keyword.IsUnknown() {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("keyword"),
				"Unnecessary Attribute",
				"keyword is only used when type is \"keyword\". It will be ignored for type \""+monType+"\".",
			)
		}
	}

	// Target format: per-type rules (URL for http/keyword/ssl; bare hostname for dns/ping; host:port for tcp)
	if !config.Target.IsNull() && !config.Target.IsUnknown() {
		for _, msg := range validateMonitorTargetFormat(monType, config.Target.ValueString()) {
			resp.Diagnostics.AddAttributeError(
				path.Root("target"),
				"Invalid Target Format",
				msg,
			)
		}
	}

	// Warn about HTTP-specific fields on non-HTTP types
	if monType == "dns" || monType == "tcp" || monType == "ping" {
		if !config.Method.IsNull() && !config.Method.IsUnknown() {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("method"),
				"Unnecessary Attribute",
				"method is only used for http and keyword monitors. It will be ignored for type \""+monType+"\".",
			)
		}
		if !config.ExpectedStatus.IsNull() && !config.ExpectedStatus.IsUnknown() {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("expected_status"),
				"Unnecessary Attribute",
				"expected_status is only used for http and keyword monitors. It will be ignored for type \""+monType+"\".",
			)
		}
	}

	// Warn about ssl_warn_days on non-SSL types
	if monType != "ssl" && monType != "" {
		if !config.SSLWarnDays.IsNull() && !config.SSLWarnDays.IsUnknown() {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("ssl_warn_days"),
				"Unnecessary Attribute",
				"ssl_warn_days is only used when type is \"ssl\". It will be ignored for type \""+monType+"\".",
			)
		}
	}
}

func (r *MonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Conversion helpers

func monitorFromModel(ctx context.Context, model MonitorResourceModel, diags *diag.Diagnostics) spork.Monitor {
	paused := model.Paused.ValueBool()
	mon := spork.Monitor{
		Target:   model.Target.ValueString(),
		Name:     model.Name.ValueString(),
		Type:     model.Type.ValueString(),
		Interval: int(model.Interval.ValueInt64()),
		Timeout:  int(model.Timeout.ValueInt64()),
		Paused:   &paused,
	}

	monType := model.Type.ValueString()

	// HTTP and keyword monitors use method, expected_status, body
	if monType == "http" || monType == "keyword" {
		if !model.Method.IsNull() && !model.Method.IsUnknown() {
			mon.Method = model.Method.ValueString()
		}
		if !model.ExpectedStatus.IsNull() && !model.ExpectedStatus.IsUnknown() {
			mon.ExpectedStatus = int(model.ExpectedStatus.ValueInt64())
		}
		if !model.Body.IsNull() && !model.Body.IsUnknown() {
			mon.Body = model.Body.ValueString()
		}
	}

	// Keyword-only fields
	if monType == "keyword" {
		if !model.Keyword.IsNull() && !model.Keyword.IsUnknown() {
			mon.Keyword = model.Keyword.ValueString()
		}
		if !model.KeywordType.IsNull() && !model.KeywordType.IsUnknown() {
			mon.KeywordType = model.KeywordType.ValueString()
		}
	}

	// SSL-only fields
	if monType == "ssl" {
		if !model.SSLWarnDays.IsNull() && !model.SSLWarnDays.IsUnknown() {
			mon.SSLWarnDays = int(model.SSLWarnDays.ValueInt64())
		}
	}

	if !model.Regions.IsNull() && !model.Regions.IsUnknown() {
		var regions []string
		diags.Append(model.Regions.ElementsAs(ctx, &regions, false)...)
		mon.Regions = regions
	}

	if !model.AlertChannelIDs.IsNull() && !model.AlertChannelIDs.IsUnknown() {
		var ids []string
		diags.Append(model.AlertChannelIDs.ElementsAs(ctx, &ids, false)...)
		mon.AlertChannelIDs = ids
	}

	if !model.Tags.IsNull() && !model.Tags.IsUnknown() {
		var tags []string
		diags.Append(model.Tags.ElementsAs(ctx, &tags, false)...)
		mon.Tags = tags
	}

	if !model.Headers.IsNull() && !model.Headers.IsUnknown() {
		var headers map[string]string
		diags.Append(model.Headers.ElementsAs(ctx, &headers, false)...)
		mon.Headers = headers
	}

	return mon
}

func monitorToModel(ctx context.Context, m spork.Monitor, diags *diag.Diagnostics) MonitorResourceModel {
	var d diag.Diagnostics
	regions, d := types.SetValueFrom(ctx, types.StringType, m.Regions)
	diags.Append(d...)
	if m.Regions == nil {
		regions, d = types.SetValueFrom(ctx, types.StringType, []string{"us-central1"})
		diags.Append(d...)
	}

	// Normalize empty slices to null so plans don't flap between
	// `alert_channel_ids = []` (from the API) and the user's
	// `alert_channel_ids = null` (the zero value for an Optional+Computed
	// attribute).
	alertChannelIDs := types.SetNull(types.StringType)
	if len(m.AlertChannelIDs) > 0 {
		alertChannelIDs, d = types.SetValueFrom(ctx, types.StringType, m.AlertChannelIDs)
		diags.Append(d...)
	}

	tags := types.SetNull(types.StringType)
	if len(m.Tags) > 0 {
		tags, d = types.SetValueFrom(ctx, types.StringType, m.Tags)
		diags.Append(d...)
	}

	headers := types.MapNull(types.StringType)
	if m.Headers != nil {
		headers, d = types.MapValueFrom(ctx, types.StringType, m.Headers)
		diags.Append(d...)
	}

	body := types.StringNull()
	if m.Body != "" {
		body = types.StringValue(m.Body)
	}

	keyword := types.StringNull()
	if m.Keyword != "" {
		keyword = types.StringValue(m.Keyword)
	}

	keywordType := types.StringNull()
	if m.KeywordType != "" {
		keywordType = types.StringValue(m.KeywordType)
	}

	sslWarnDays := types.Int64Null()
	if m.SSLWarnDays != 0 {
		sslWarnDays = types.Int64Value(int64(m.SSLWarnDays))
	}

	method := types.StringNull()
	if m.Method != "" {
		method = types.StringValue(m.Method)
	}

	expectedStatus := types.Int64Null()
	if m.ExpectedStatus != 0 {
		expectedStatus = types.Int64Value(int64(m.ExpectedStatus))
	}

	return MonitorResourceModel{
		ID:              types.StringValue(m.ID),
		Target:          types.StringValue(m.Target),
		Name:            types.StringValue(m.Name),
		Type:            types.StringValue(m.Type),
		Method:          method,
		ExpectedStatus:  expectedStatus,
		Interval:        types.Int64Value(int64(m.Interval)),
		Timeout:         types.Int64Value(int64(m.Timeout)),
		Regions:         regions,
		AlertChannelIDs: alertChannelIDs,
		Tags:            tags,
		Paused:          types.BoolValue(m.Paused != nil && *m.Paused),
		Status:          types.StringValue(m.Status),
		Headers:         headers,
		Body:            body,
		Keyword:         keyword,
		KeywordType:     keywordType,
		SSLWarnDays:     sslWarnDays,
	}
}

// validateMonitorTargetFormat returns human-readable error messages for any
// target format violations given the monitor type. An empty slice means the
// target is valid. The rules mirror the server's per-type expectations:
//
//	http, keyword, ssl: URL starting with http:// or https://
//	dns, ping:          bare hostname or IP — no scheme, no path, no port
//	tcp:                host:port — no scheme, no path; port must be 1-65535
//
// Unknown monType (e.g. empty string when type uses a variable) returns no
// errors; the type validator handles that separately.
func validateMonitorTargetFormat(monType, target string) []string {
	if target == "" {
		return nil
	}

	switch monType {
	case "http", "keyword", "ssl":
		if !hasURLScheme(target) {
			return []string{"target must start with http:// or https:// for monitor type \"" + monType + "\"."}
		}
		return nil

	case "dns", "ping":
		var issues []string
		if hasURLScheme(target) {
			issues = append(issues, "target for \""+monType+"\" monitors must be a bare hostname or IP (e.g., example.com) — URL schemes are not allowed.")
		}
		if strings.Contains(target, "/") {
			issues = append(issues, "target for \""+monType+"\" monitors must not include a path.")
		}
		// Ports aren't meaningful for DNS resolution or ICMP. If SplitHostPort
		// succeeds, the user included a port — reject it.
		if _, _, err := net.SplitHostPort(target); err == nil {
			issues = append(issues, "target for \""+monType+"\" monitors must not include a port.")
		}
		return issues

	case "tcp":
		var issues []string
		if hasURLScheme(target) {
			issues = append(issues, "target for \"tcp\" monitors must be host:port (e.g., example.com:443) — URL schemes are not allowed.")
			return issues
		}
		if strings.Contains(target, "/") {
			issues = append(issues, "target for \"tcp\" monitors must not include a path.")
			return issues
		}
		host, portStr, err := net.SplitHostPort(target)
		if err != nil {
			return []string{"target for \"tcp\" monitors must be host:port (e.g., example.com:443)."}
		}
		if host == "" {
			issues = append(issues, "target for \"tcp\" monitors must include a host before the port.")
		}
		port, perr := strconv.Atoi(portStr)
		if perr != nil || port < 1 || port > 65535 {
			issues = append(issues, fmt.Sprintf("target port %q for \"tcp\" monitors must be an integer between 1 and 65535.", portStr))
		}
		return issues
	}

	return nil
}

// hasURLScheme reports whether target starts with "http://" or "https://"
// (case-insensitive).
func hasURLScheme(target string) bool {
	lower := strings.ToLower(target)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}
