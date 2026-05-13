package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	spork "github.com/sporkops/spork-go"
)

// Plan-time same-org check for sporkops_status_page.components[].monitor_id.
// The provider is org-scoped via spork.WithOrganization; a GetMonitor that
// 404s means "not in this org" (whether the ID actually exists in another
// org or doesn't exist at all is intentionally indistinguishable, mirroring
// the backend's privacy posture).

// pingMonitorStub returns a tiny Ping API that knows about exactly the
// monitor IDs in `owned`. Anything else 404s. Mirrors the backend route:
// GET /v1/orgs/{orgID}/monitors/{id}.
func pingMonitorStub(t *testing.T, owned ...string) *httptest.Server {
	t.Helper()
	known := make(map[string]struct{}, len(owned))
	for _, id := range owned {
		known[id] = struct{}{}
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// /v1/orgs/{orgID}/monitors/{id}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 5 || parts[0] != "v1" || parts[1] != "orgs" || parts[3] != "monitors" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		id := parts[4]
		if _, ok := known[id]; !ok {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{"code": "not_found", "message": "monitor not found"},
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"id": id, "organization_id": parts[2]},
		})
	}))
}

func newClientForStub(t *testing.T, srv *httptest.Server) *spork.Client {
	t.Helper()
	return spork.NewClient(
		spork.WithAPIKey("sk_test"),
		spork.WithOrganization("org-acme"),
		spork.WithBaseURL(srv.URL+"/v1"),
	)
}

func TestVerifyComponentMonitorsSameOrg_AllOwned_NoDiagnostics(t *testing.T) {
	srv := pingMonitorStub(t, "mon-owned-1", "mon-owned-2")
	defer srv.Close()
	client := newClientForStub(t, srv)

	components := []StatusPageComponentModel{
		{MonitorID: types.StringValue("mon-owned-1")},
		{MonitorID: types.StringValue("mon-owned-2")},
	}
	diags := verifyComponentMonitorsSameOrg(context.Background(), client, components)
	if diags.HasError() {
		t.Fatalf("expected no errors, got: %s", diags.Errors())
	}
	if len(diags.Warnings()) != 0 {
		t.Fatalf("expected no warnings, got: %s", diags.Warnings())
	}
}

func TestVerifyComponentMonitorsSameOrg_OneInDifferentOrg_PlanError(t *testing.T) {
	srv := pingMonitorStub(t, "mon-owned")
	defer srv.Close()
	client := newClientForStub(t, srv)

	components := []StatusPageComponentModel{
		{MonitorID: types.StringValue("mon-owned")},
		{MonitorID: types.StringValue("mon-elsewhere")},
	}
	diags := verifyComponentMonitorsSameOrg(context.Background(), client, components)
	if !diags.HasError() {
		t.Fatal("expected an error diagnostic for the cross-org monitor")
	}
	if len(diags.Errors()) != 1 {
		t.Fatalf("expected exactly 1 error, got %d: %s", len(diags.Errors()), diags.Errors())
	}
	errDetail := diags.Errors()[0].Detail()
	if !strings.Contains(errDetail, "mon-elsewhere") {
		t.Errorf("expected error to name the offending ID 'mon-elsewhere', got: %s", errDetail)
	}
	if !strings.Contains(errDetail, "provider = spork.") {
		t.Errorf("expected error to hint at provider alias fix, got: %s", errDetail)
	}
}

func TestVerifyComponentMonitorsSameOrg_UnknownAndNull_Skipped(t *testing.T) {
	// If the API is called, the stub will 404 (it owns no monitors). The
	// test passes only if verifyComponentMonitorsSameOrg never reaches it.
	srv := pingMonitorStub(t)
	defer srv.Close()
	client := newClientForStub(t, srv)

	components := []StatusPageComponentModel{
		{MonitorID: types.StringUnknown()},
		{MonitorID: types.StringNull()},
		{MonitorID: types.StringValue("")}, // empty after .ValueString() — also skipped
	}
	diags := verifyComponentMonitorsSameOrg(context.Background(), client, components)
	if diags.HasError() {
		t.Fatalf("expected no errors for unknown/null/empty monitor_ids, got: %s", diags.Errors())
	}
	if len(diags.Warnings()) != 0 {
		t.Fatalf("expected no warnings, got: %s", diags.Warnings())
	}
}

func TestVerifyComponentMonitorsSameOrg_DuplicateIDs_OneAPICall(t *testing.T) {
	// Stub counts calls so we can assert on dedup.
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"id": "mon-owned", "organization_id": "org-acme"},
		})
	}))
	defer srv.Close()
	client := newClientForStub(t, srv)

	components := []StatusPageComponentModel{
		{MonitorID: types.StringValue("mon-owned")},
		{MonitorID: types.StringValue("mon-owned")},
		{MonitorID: types.StringValue("mon-owned")},
	}
	diags := verifyComponentMonitorsSameOrg(context.Background(), client, components)
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags.Errors())
	}
	if calls != 1 {
		t.Fatalf("expected exactly 1 API call (dedup), got %d", calls)
	}
}

func TestVerifyComponentMonitorsSameOrg_TransientError_WarningNotError(t *testing.T) {
	// 500 from the stub mimics a transient backend issue. The check must
	// surface a Warning so the plan can proceed; the apply will hit the
	// real backend and either succeed or surface the canonical error.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"code":"internal","message":"boom"}}`))
	}))
	defer srv.Close()
	client := newClientForStub(t, srv)

	components := []StatusPageComponentModel{
		{MonitorID: types.StringValue("mon-anything")},
	}
	diags := verifyComponentMonitorsSameOrg(context.Background(), client, components)
	if diags.HasError() {
		t.Fatalf("transient errors must not block the plan, got error: %s", diags.Errors())
	}
	if len(diags.Warnings()) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(diags.Warnings()), diags.Warnings())
	}
}
