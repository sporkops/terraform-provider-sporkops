package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	spork "github.com/sporkops/spork-go"
)

func TestMonitorFromModel_dns(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := MonitorResourceModel{
		Target:   types.StringValue("sporkops.com"),
		Name:     types.StringValue("DNS Test"),
		Type:     types.StringValue("dns"),
		Interval: types.Int64Value(300),
		Timeout:  types.Int64Value(30),
		Paused:   types.BoolValue(false),
		// HTTP-specific fields are null (not set by user, no schema default)
		Method:         types.StringNull(),
		ExpectedStatus: types.Int64Null(),
		Body:           types.StringNull(),
		Keyword:        types.StringNull(),
		KeywordType:    types.StringNull(),
		SSLWarnDays:    types.Int64Null(),
		Regions:        types.ListNull(types.StringType),
		AlertChannelIDs: types.ListNull(types.StringType),
		Tags:           types.ListNull(types.StringType),
		Headers:        types.MapNull(types.StringType),
	}

	mon := monitorFromModel(ctx, model, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	if mon.Type != "dns" {
		t.Errorf("Type = %q, want %q", mon.Type, "dns")
	}
	if mon.Method != "" {
		t.Errorf("Method = %q, want empty (should not be set for dns)", mon.Method)
	}
	if mon.ExpectedStatus != 0 {
		t.Errorf("ExpectedStatus = %d, want 0 (should not be set for dns)", mon.ExpectedStatus)
	}
	if mon.KeywordType != "" {
		t.Errorf("KeywordType = %q, want empty (should not be set for dns)", mon.KeywordType)
	}
	if mon.SSLWarnDays != 0 {
		t.Errorf("SSLWarnDays = %d, want 0 (should not be set for dns)", mon.SSLWarnDays)
	}
	if mon.Body != "" {
		t.Errorf("Body = %q, want empty (should not be set for dns)", mon.Body)
	}

	// Verify JSON marshaling omits these fields
	data, err := json.Marshal(mon)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	var raw map[string]any
	json.Unmarshal(data, &raw)
	for _, key := range []string{"method", "expected_status", "keyword_type", "ssl_warn_days", "body", "keyword"} {
		if _, ok := raw[key]; ok {
			t.Errorf("JSON should not contain %q for dns monitor, but it does", key)
		}
	}
}

func TestMonitorFromModel_http(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := MonitorResourceModel{
		Target:         types.StringValue("https://example.com"),
		Name:           types.StringValue("HTTP Test"),
		Type:           types.StringValue("http"),
		Method:         types.StringValue("POST"),
		ExpectedStatus: types.Int64Value(201),
		Interval:       types.Int64Value(60),
		Timeout:        types.Int64Value(30),
		Paused:         types.BoolValue(false),
		Body:           types.StringValue(`{"key":"value"}`),
		Keyword:        types.StringNull(),
		KeywordType:    types.StringNull(),
		SSLWarnDays:    types.Int64Null(),
		Regions:        types.ListNull(types.StringType),
		AlertChannelIDs: types.ListNull(types.StringType),
		Tags:           types.ListNull(types.StringType),
		Headers:        types.MapNull(types.StringType),
	}

	mon := monitorFromModel(ctx, model, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	if mon.Method != "POST" {
		t.Errorf("Method = %q, want %q", mon.Method, "POST")
	}
	if mon.ExpectedStatus != 201 {
		t.Errorf("ExpectedStatus = %d, want %d", mon.ExpectedStatus, 201)
	}
	if mon.Body != `{"key":"value"}` {
		t.Errorf("Body = %q, want %q", mon.Body, `{"key":"value"}`)
	}
}

func TestMonitorFromModel_ssl(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := MonitorResourceModel{
		Target:         types.StringValue("https://example.com"),
		Name:           types.StringValue("SSL Test"),
		Type:           types.StringValue("ssl"),
		Interval:       types.Int64Value(300),
		Timeout:        types.Int64Value(30),
		Paused:         types.BoolValue(false),
		SSLWarnDays:    types.Int64Value(14),
		Method:         types.StringNull(),
		ExpectedStatus: types.Int64Null(),
		Body:           types.StringNull(),
		Keyword:        types.StringNull(),
		KeywordType:    types.StringNull(),
		Regions:        types.ListNull(types.StringType),
		AlertChannelIDs: types.ListNull(types.StringType),
		Tags:           types.ListNull(types.StringType),
		Headers:        types.MapNull(types.StringType),
	}

	mon := monitorFromModel(ctx, model, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	if mon.SSLWarnDays != 14 {
		t.Errorf("SSLWarnDays = %d, want %d", mon.SSLWarnDays, 14)
	}
	if mon.Method != "" {
		t.Errorf("Method = %q, want empty (should not be set for ssl)", mon.Method)
	}
	if mon.ExpectedStatus != 0 {
		t.Errorf("ExpectedStatus = %d, want 0 (should not be set for ssl)", mon.ExpectedStatus)
	}
}

func TestMonitorFromModel_keyword(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := MonitorResourceModel{
		Target:         types.StringValue("https://example.com/health"),
		Name:           types.StringValue("Keyword Test"),
		Type:           types.StringValue("keyword"),
		Method:         types.StringValue("GET"),
		ExpectedStatus: types.Int64Value(200),
		Interval:       types.Int64Value(60),
		Timeout:        types.Int64Value(30),
		Paused:         types.BoolValue(false),
		Keyword:        types.StringValue("healthy"),
		KeywordType:    types.StringValue("exists"),
		SSLWarnDays:    types.Int64Null(),
		Body:           types.StringNull(),
		Regions:        types.ListNull(types.StringType),
		AlertChannelIDs: types.ListNull(types.StringType),
		Tags:           types.ListNull(types.StringType),
		Headers:        types.MapNull(types.StringType),
	}

	mon := monitorFromModel(ctx, model, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	if mon.Method != "GET" {
		t.Errorf("Method = %q, want %q", mon.Method, "GET")
	}
	if mon.Keyword != "healthy" {
		t.Errorf("Keyword = %q, want %q", mon.Keyword, "healthy")
	}
	if mon.KeywordType != "exists" {
		t.Errorf("KeywordType = %q, want %q", mon.KeywordType, "exists")
	}
	if mon.SSLWarnDays != 0 {
		t.Errorf("SSLWarnDays = %d, want 0 (should not be set for keyword)", mon.SSLWarnDays)
	}
}

func TestMonitorFromModel_tcp(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := MonitorResourceModel{
		Target:         types.StringValue("sporkops.com:443"),
		Name:           types.StringValue("TCP Test"),
		Type:           types.StringValue("tcp"),
		Interval:       types.Int64Value(120),
		Timeout:        types.Int64Value(30),
		Paused:         types.BoolValue(false),
		Method:         types.StringNull(),
		ExpectedStatus: types.Int64Null(),
		Body:           types.StringNull(),
		Keyword:        types.StringNull(),
		KeywordType:    types.StringNull(),
		SSLWarnDays:    types.Int64Null(),
		Regions:        types.ListNull(types.StringType),
		AlertChannelIDs: types.ListNull(types.StringType),
		Tags:           types.ListNull(types.StringType),
		Headers:        types.MapNull(types.StringType),
	}

	mon := monitorFromModel(ctx, model, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	if mon.Method != "" {
		t.Errorf("Method = %q, want empty", mon.Method)
	}
	if mon.ExpectedStatus != 0 {
		t.Errorf("ExpectedStatus = %d, want 0", mon.ExpectedStatus)
	}
}

func TestMonitorFromModel_ping(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := MonitorResourceModel{
		Target:         types.StringValue("sporkops.com"),
		Name:           types.StringValue("Ping Test"),
		Type:           types.StringValue("ping"),
		Interval:       types.Int64Value(120),
		Timeout:        types.Int64Value(30),
		Paused:         types.BoolValue(false),
		Method:         types.StringNull(),
		ExpectedStatus: types.Int64Null(),
		Body:           types.StringNull(),
		Keyword:        types.StringNull(),
		KeywordType:    types.StringNull(),
		SSLWarnDays:    types.Int64Null(),
		Regions:        types.ListNull(types.StringType),
		AlertChannelIDs: types.ListNull(types.StringType),
		Tags:           types.ListNull(types.StringType),
		Headers:        types.MapNull(types.StringType),
	}

	mon := monitorFromModel(ctx, model, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	if mon.Method != "" {
		t.Errorf("Method = %q, want empty", mon.Method)
	}
	if mon.ExpectedStatus != 0 {
		t.Errorf("ExpectedStatus = %d, want 0", mon.ExpectedStatus)
	}
}

func TestMonitorToModel_dns(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Simulate API response for a DNS monitor (no HTTP-specific fields)
	apiMonitor := spork.Monitor{
		ID:       "mon_123",
		Name:     "DNS Test",
		Type:     "dns",
		Target:   "sporkops.com",
		Interval: 300,
		Timeout:  30,
		Status:   "up",
		// Method, ExpectedStatus, KeywordType, SSLWarnDays are zero values
	}

	model := monitorToModel(ctx, apiMonitor, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	if !model.Method.IsNull() {
		t.Errorf("Method should be null for dns monitor, got %q", model.Method.ValueString())
	}
	if !model.ExpectedStatus.IsNull() {
		t.Errorf("ExpectedStatus should be null for dns monitor, got %d", model.ExpectedStatus.ValueInt64())
	}
	if !model.KeywordType.IsNull() {
		t.Errorf("KeywordType should be null for dns monitor, got %q", model.KeywordType.ValueString())
	}
	if !model.SSLWarnDays.IsNull() {
		t.Errorf("SSLWarnDays should be null for dns monitor, got %d", model.SSLWarnDays.ValueInt64())
	}
}

func TestMonitorToModel_http(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Simulate API response for an HTTP monitor
	apiMonitor := spork.Monitor{
		ID:             "mon_456",
		Name:           "HTTP Test",
		Type:           "http",
		Target:         "https://example.com",
		Method:         "GET",
		ExpectedStatus: 200,
		Interval:       60,
		Timeout:        30,
		Status:         "up",
	}

	model := monitorToModel(ctx, apiMonitor, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	if model.Method.IsNull() || model.Method.ValueString() != "GET" {
		t.Errorf("Method = %v, want %q", model.Method, "GET")
	}
	if model.ExpectedStatus.IsNull() || model.ExpectedStatus.ValueInt64() != 200 {
		t.Errorf("ExpectedStatus = %v, want %d", model.ExpectedStatus, 200)
	}
}
