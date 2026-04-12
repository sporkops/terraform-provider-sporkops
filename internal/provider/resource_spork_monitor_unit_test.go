package provider

import (
	"context"
	"encoding/json"
	"strings"
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
		Regions:        types.SetNull(types.StringType),
		AlertChannelIDs: types.SetNull(types.StringType),
		Tags:           types.SetNull(types.StringType),
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
		Regions:        types.SetNull(types.StringType),
		AlertChannelIDs: types.SetNull(types.StringType),
		Tags:           types.SetNull(types.StringType),
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
		Regions:        types.SetNull(types.StringType),
		AlertChannelIDs: types.SetNull(types.StringType),
		Tags:           types.SetNull(types.StringType),
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
		Regions:        types.SetNull(types.StringType),
		AlertChannelIDs: types.SetNull(types.StringType),
		Tags:           types.SetNull(types.StringType),
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
		Regions:        types.SetNull(types.StringType),
		AlertChannelIDs: types.SetNull(types.StringType),
		Tags:           types.SetNull(types.StringType),
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
		Regions:        types.SetNull(types.StringType),
		AlertChannelIDs: types.SetNull(types.StringType),
		Tags:           types.SetNull(types.StringType),
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

func TestValidateMonitorTargetFormat(t *testing.T) {
	tests := []struct {
		name         string
		monType      string
		target       string
		wantErr      bool
		wantContains string // substring expected in first error message, if wantErr
	}{
		// http / keyword / ssl — require URL scheme
		{"http ok", "http", "https://example.com", false, ""},
		{"http ok plain http", "http", "http://example.com", false, ""},
		{"http missing scheme", "http", "example.com", true, "http://"},
		{"keyword ok", "keyword", "https://example.com/health", false, ""},
		{"keyword missing scheme", "keyword", "example.com", true, "http://"},
		{"ssl ok", "ssl", "https://example.com", false, ""},
		{"ssl missing scheme", "ssl", "example.com", true, "http://"},

		// dns / ping — bare hostname only
		{"dns ok hostname", "dns", "example.com", false, ""},
		{"dns ok ip", "dns", "8.8.8.8", false, ""},
		{"dns rejects https scheme", "dns", "https://example.com", true, "URL schemes are not allowed"},
		{"dns rejects http scheme", "dns", "http://example.com", true, "URL schemes are not allowed"},
		{"dns rejects mixed case scheme", "dns", "HTTPS://example.com", true, "URL schemes are not allowed"},
		{"dns rejects path", "dns", "example.com/path", true, "must not include a path"},
		{"dns rejects port", "dns", "example.com:443", true, "must not include a port"},
		{"ping ok", "ping", "example.com", false, ""},
		{"ping rejects scheme", "ping", "http://example.com", true, "URL schemes are not allowed"},
		{"ping rejects port", "ping", "example.com:80", true, "must not include a port"},

		// tcp — host:port required
		{"tcp ok", "tcp", "example.com:443", false, ""},
		{"tcp ok low port", "tcp", "example.com:1", false, ""},
		{"tcp ok high port", "tcp", "example.com:65535", false, ""},
		{"tcp ok ipv4", "tcp", "10.0.0.1:22", false, ""},
		{"tcp missing port", "tcp", "example.com", true, "host:port"},
		{"tcp rejects scheme", "tcp", "https://example.com:443", true, "URL schemes are not allowed"},
		{"tcp rejects path", "tcp", "example.com:443/x", true, "path"},
		{"tcp rejects port zero", "tcp", "example.com:0", true, "1 and 65535"},
		{"tcp rejects port too high", "tcp", "example.com:99999", true, "1 and 65535"},
		{"tcp rejects non-numeric port", "tcp", "example.com:abc", true, "1 and 65535"},
		{"tcp rejects empty host", "tcp", ":443", true, "host"},

		// Edge cases
		{"empty target is valid at this layer", "dns", "", false, ""}, // Required-ness is enforced elsewhere
		{"unknown type no-op", "", "whatever", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := validateMonitorTargetFormat(tt.monType, tt.target)
			gotErr := len(issues) > 0
			if gotErr != tt.wantErr {
				t.Fatalf("wantErr=%v, got issues=%v", tt.wantErr, issues)
			}
			if tt.wantErr && tt.wantContains != "" {
				found := false
				for _, msg := range issues {
					if strings.Contains(msg, tt.wantContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected an issue containing %q, got: %v", tt.wantContains, issues)
				}
			}
		})
	}
}
