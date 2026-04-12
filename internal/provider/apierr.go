package provider

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/sporkops/spork-go"
)

// addAPIError appends a diagnostic that preserves the structured information
// carried by a *spork.APIError (HTTP status, API error code, field-level
// validation details, and the X-Request-Id returned by the server). For
// non-API errors the original error string is surfaced unchanged.
//
// The caller-supplied summary is the short title Terraform shows at the top
// of the diagnostic (e.g. "Error creating monitor"); it is left alone so
// existing call sites do not need to re-learn their phrasing. For API errors
// the summary is additionally annotated with the HTTP status code so
// operators can distinguish 4xx user errors from 5xx server errors at a
// glance.
func addAPIError(diags *diag.Diagnostics, summary string, err error) {
	if err == nil {
		return
	}

	var apiErr *spork.APIError
	if !errors.As(err, &apiErr) {
		diags.AddError(summary, err.Error())
		return
	}

	// Annotate the summary with the status code for quick triage.
	if apiErr.StatusCode > 0 {
		summary = fmt.Sprintf("%s (HTTP %d)", summary, apiErr.StatusCode)
	}

	var b strings.Builder
	if apiErr.Message != "" {
		b.WriteString(apiErr.Message)
	}
	if apiErr.Code != "" && apiErr.Code != "unknown" {
		fmt.Fprintf(&b, "\n\nError code: %s", apiErr.Code)
	}
	if len(apiErr.Details) > 0 {
		b.WriteString("\n\nField errors:")
		for _, d := range apiErr.Details {
			if d.Field != "" {
				fmt.Fprintf(&b, "\n  - %s: %s", d.Field, d.Message)
			} else {
				fmt.Fprintf(&b, "\n  - %s", d.Message)
			}
		}
	}
	if apiErr.RequestID != "" {
		fmt.Fprintf(&b, "\n\nRequest ID: %s (include this when contacting support)", apiErr.RequestID)
	}

	diags.AddError(summary, b.String())
}
