package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	spork "github.com/sporkops/spork-go"
)

// Org-qualified import support.
//
// Terraform's default `terraform import RESOURCE.NAME ID` flow makes
// every imported resource read from whichever org the provider alias
// is configured for. That works fine in single-org configs but is
// awkward in multi-org setups: importing a resource owned by org B
// into a config wired to provider alias for org A would either need a
// per-org alias or fail at read time.
//
// We accept a second import form — `ORG_ID:RESOURCE_ID` — which pins
// the resource's tenancy at import time. The org gets stored in the
// resource's `organization_id` attribute and every subsequent
// Read / Update / Delete routes through Client.ForOrg(state.OrgID),
// so the resource stays addressable regardless of the provider
// alias's configured org. The legacy single-arg form still works
// unchanged.
//
// This is the same pattern Stripe's TF provider uses for connected
// accounts (`acct_id:resource_id`) and what the AWS provider uses for
// cross-region imports. It pairs naturally with the SDK's ForOrg
// helper, which is a shallow-clone returning a new *spork.Client
// scoped to the supplied org — no per-org connection cost.

// parseImportID splits an import ID into (orgID, resourceID). When the
// caller used the legacy single-arg form, orgID is empty and the
// resource will use the provider's configured org for all calls.
//
// Whitespace on either side of the colon is tolerated so a copy-paste
// from a doc or chat doesn't break the import.
func parseImportID(raw string) (orgID, resourceID string) {
	raw = strings.TrimSpace(raw)
	if i := strings.IndexByte(raw, ':'); i >= 0 {
		return strings.TrimSpace(raw[:i]), strings.TrimSpace(raw[i+1:])
	}
	return "", raw
}

// handleOrgQualifiedImport is the framework-side adapter — each
// resource's ImportState method delegates to this helper so they all
// share the same parsing and error message. Sets `id` always and
// `organization_id` only when the import was org-qualified; the
// latter being null is the "use the provider's org" signal that
// clientForState reads back during Read / Update / Delete.
func handleOrgQualifiedImport(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	orgID, resourceID := parseImportID(req.ID)
	if resourceID == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("expected `RESOURCE_ID` or `ORG_ID:RESOURCE_ID`, got %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), resourceID)...)
	if orgID != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	}
}

// clientForState returns the *spork.Client to use for an org-scoped
// request. When the resource carries an explicit organization_id in
// state — either set at create time from the provider's resolved
// org, or pinned by an org-qualified import — the call is routed
// through Client.ForOrg so it lands on the right tenant regardless
// of which org the provider was configured for at startup.
//
// Returns the original client when organization_id is null / unknown
// / empty so behaviour stays unchanged for the dominant case of a
// single-org provider config that never sees the org-qualified import
// form.
func clientForState(c *spork.Client, orgID types.String) *spork.Client {
	if c == nil {
		return nil
	}
	if orgID.IsNull() || orgID.IsUnknown() {
		return c
	}
	id := strings.TrimSpace(orgID.ValueString())
	if id == "" {
		return c
	}
	return c.ForOrg(id)
}

// resolveCreateOrg returns the org ID to record on a freshly-created
// resource. We ask the SDK (rather than the provider config) so the
// stored value matches whichever org the request actually hit — the
// provider may have been configured without an explicit
// organization_id and let the SDK auto-resolve on first call.
//
// Errors are non-fatal at the call site: a resource created without
// an organization_id in state still works (it falls back to the
// provider's client on every subsequent call), but Terraform users
// won't be able to address it independently of the provider alias.
func resolveCreateOrg(ctx context.Context, c *spork.Client) (string, error) {
	if c == nil {
		return "", fmt.Errorf("client not configured")
	}
	return c.OrganizationID(ctx)
}
