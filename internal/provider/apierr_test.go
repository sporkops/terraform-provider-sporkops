package provider

import (
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/sporkops/spork-go"
)

func TestAddAPIError_WrapsStructuredAPIError(t *testing.T) {
	var diags diag.Diagnostics
	apiErr := &spork.APIError{
		StatusCode: 422,
		Code:       "validation_error",
		Message:    "target is required",
		RequestID:  "req_abc123",
	}
	addAPIError(&diags, "Error creating monitor", apiErr)

	if diags.ErrorsCount() != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", diags.ErrorsCount())
	}
	d := diags.Errors()[0]
	summary := d.Summary()
	detail := d.Detail()

	if !strings.Contains(summary, "HTTP 422") {
		t.Errorf("expected summary to include HTTP status, got %q", summary)
	}
	if !strings.Contains(summary, "Error creating monitor") {
		t.Errorf("expected summary to preserve caller prefix, got %q", summary)
	}
	if !strings.Contains(detail, "target is required") {
		t.Errorf("expected detail to include API message, got %q", detail)
	}
	if !strings.Contains(detail, "validation_error") {
		t.Errorf("expected detail to include error code, got %q", detail)
	}
	if !strings.Contains(detail, "req_abc123") {
		t.Errorf("expected detail to include request id, got %q", detail)
	}
}

func TestAddAPIError_PassesThroughNonAPIError(t *testing.T) {
	var diags diag.Diagnostics
	addAPIError(&diags, "Network failure", errors.New("connection refused"))

	if diags.ErrorsCount() != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", diags.ErrorsCount())
	}
	d := diags.Errors()[0]
	if d.Summary() != "Network failure" {
		t.Errorf("expected raw summary, got %q", d.Summary())
	}
	if d.Detail() != "connection refused" {
		t.Errorf("expected raw error in detail, got %q", d.Detail())
	}
}

func TestAddAPIError_NilIsNoOp(t *testing.T) {
	var diags diag.Diagnostics
	addAPIError(&diags, "Should not appear", nil)
	if diags.ErrorsCount() != 0 {
		t.Fatalf("expected 0 diagnostics for nil err, got %d", diags.ErrorsCount())
	}
}

func TestAddAPIError_SurfacesFieldDetails(t *testing.T) {
	var diags diag.Diagnostics
	apiErr := &spork.APIError{
		StatusCode: 422,
		Code:       "validation_error",
		Message:    "invalid monitor",
		Details: []spork.ErrorDetail{
			{Field: "target", Message: "must be a valid URL"},
			{Field: "interval", Message: "must be >= 60"},
			{Message: "one or more fields failed validation"},
		},
		RequestID: "req_xyz",
	}
	addAPIError(&diags, "Error creating monitor", apiErr)

	d := diags.Errors()[0]
	detail := d.Detail()
	if !strings.Contains(detail, "Field errors:") {
		t.Errorf("expected Field errors: block, got %q", detail)
	}
	if !strings.Contains(detail, "target: must be a valid URL") {
		t.Errorf("expected target detail, got %q", detail)
	}
	if !strings.Contains(detail, "interval: must be >= 60") {
		t.Errorf("expected interval detail, got %q", detail)
	}
	if !strings.Contains(detail, "one or more fields failed validation") {
		t.Errorf("expected fieldless detail to appear, got %q", detail)
	}
}

func TestAddAPIError_SkipsUnknownCodeAndEmptyRequestID(t *testing.T) {
	var diags diag.Diagnostics
	apiErr := &spork.APIError{
		StatusCode: 500,
		Code:       "unknown",
		Message:    "internal server error",
	}
	addAPIError(&diags, "Error reading monitor", apiErr)

	d := diags.Errors()[0]
	if strings.Contains(d.Detail(), "Error code:") {
		t.Errorf("expected detail to omit unknown error code, got %q", d.Detail())
	}
	if strings.Contains(d.Detail(), "Request ID:") {
		t.Errorf("expected detail to omit empty request id, got %q", d.Detail())
	}
	if !strings.Contains(d.Summary(), "HTTP 500") {
		t.Errorf("expected summary to include HTTP 500, got %q", d.Summary())
	}
}
