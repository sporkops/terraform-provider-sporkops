package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	spork "github.com/sporkops/spork-go"
)

// parseImportID accepts two forms — `RESOURCE_ID` and `ORG_ID:RESOURCE_ID` —
// and is consumed by every resource's ImportState method. These tests pin
// the parsing contract: the colon split, the whitespace tolerance, and
// the empty-result behaviour that drives the "invalid import ID" error.

func TestParseImportID(t *testing.T) {
	cases := []struct {
		name        string
		raw         string
		wantOrg     string
		wantResrc   string
	}{
		{"legacy bare ID", "mon_abc", "", "mon_abc"},
		{"org-qualified", "org_xyz:mon_abc", "org_xyz", "mon_abc"},
		{"whitespace tolerated", "  org_xyz : mon_abc  ", "org_xyz", "mon_abc"},
		{"trailing colon gives empty resource", "org_xyz:", "org_xyz", ""},
		{"leading colon gives empty org", ":mon_abc", "", "mon_abc"},
		{"completely empty", "", "", ""},
		{"only colon", ":", "", ""},
		// Resource IDs containing colons (none exist today, but the
		// helper should split on the *first* colon so a future "ID with
		// embedded :" doesn't accidentally claim the prefix as org_id.
		{"first colon wins", "org_xyz:wat:ever", "org_xyz", "wat:ever"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotOrg, gotResrc := parseImportID(tc.raw)
			if gotOrg != tc.wantOrg {
				t.Errorf("org = %q, want %q", gotOrg, tc.wantOrg)
			}
			if gotResrc != tc.wantResrc {
				t.Errorf("resource = %q, want %q", gotResrc, tc.wantResrc)
			}
		})
	}
}

// clientForState selects between the provider's configured client and a
// per-org ForOrg clone based on the resource's state. The path is hit on
// every Read / Update / Delete, so the no-op fast path for the dominant
// (single-org, no org_id in state) case has to be exact.

func TestClientForState_NullOrEmptyReturnsOriginal(t *testing.T) {
	c := spork.NewClient(
		spork.WithAPIKey("sk_test"),
		spork.WithOrganization("org_provider"),
	)
	cases := map[string]types.String{
		"null":    types.StringNull(),
		"unknown": types.StringUnknown(),
		"empty":   types.StringValue(""),
		"spaces":  types.StringValue("   "),
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			if got := clientForState(c, in); got != c {
				t.Errorf("expected original client back, got a different pointer")
			}
		})
	}
}

func TestClientForState_NonEmptyRoutesThroughForOrg(t *testing.T) {
	// We don't have a public accessor for the client's resolved org ID
	// in the package, so verify the routing by observing the URL the
	// next request goes to. The stub returns 404 for everything; we
	// only need it to receive a request to inspect the path.
	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Path
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"code":"not_found"}}`))
	}))
	defer srv.Close()

	c := spork.NewClient(
		spork.WithAPIKey("sk_test"),
		spork.WithOrganization("org_provider"),
		spork.WithBaseURL(srv.URL+"/v1"),
	)

	scoped := clientForState(c, types.StringValue("org_other"))
	if scoped == c {
		t.Fatal("expected a per-call clone, got the original client")
	}
	// Any GetMonitor call will land on /orgs/{org}/monitors/{id}; the
	// 404 from the stub is fine — we just care which org was in the URL.
	_, _ = scoped.GetMonitor(context.Background(), "mon_x")
	if seen == "" {
		t.Fatal("stub never received a request")
	}
	if want := "/v1/orgs/org_other/monitors/mon_x"; seen != want {
		t.Errorf("path = %q, want %q (ForOrg should override the provider's org)", seen, want)
	}
}

func TestClientForState_NilClientReturnsNil(t *testing.T) {
	// Defensive: framework probes can land here before Configure runs.
	// Returning nil keeps the helper safe to call unconditionally; the
	// real Read path bails earlier via the resource's own client check.
	if got := clientForState(nil, types.StringValue("org_x")); got != nil {
		t.Errorf("expected nil for nil client, got %v", got)
	}
}
