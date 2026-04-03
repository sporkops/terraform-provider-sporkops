package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	spork "github.com/sporkops/spork-go"
)

func TestAlertChannelToModel_preservesRedactedConfig(t *testing.T) {
	ctx := context.Background()
	fullURL := "https://hooks.example.com/services/T00/B00/xxxxxxxxxxxxx"

	tests := []struct {
		name       string
		apiConfig  map[string]string // what the API returns
		wantURL    string            // expected url in resulting model
	}{
		{
			name:      "masked URL with ellipsis",
			apiConfig: map[string]string{"url": "https://h...m/xxxxxxxxxxxxx"},
			wantURL:   fullURL,
		},
		{
			name:      "empty URL",
			apiConfig: map[string]string{"url": ""},
			wantURL:   fullURL,
		},
		{
			name:      "missing URL key",
			apiConfig: map[string]string{},
			wantURL:   fullURL,
		},
		{
			name:      "truncated URL shorter than original",
			apiConfig: map[string]string{"url": "https://hooks.e*****"},
			wantURL:   fullURL,
		},
		{
			name:      "full URL returned unchanged",
			apiConfig: map[string]string{"url": fullURL},
			wantURL:   fullURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiChannel := spork.AlertChannel{
				ID:       "ch_123",
				Name:     "Test Webhook",
				Type:     "webhook",
				Verified: true,
				Config:   tt.apiConfig,
			}

			fallbackConfig, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{
				"url": fullURL,
			})
			fallback := &AlertChannelResourceModel{
				ID:     types.StringValue("ch_123"),
				Name:   types.StringValue("Test Webhook"),
				Type:   types.StringValue("webhook"),
				Config: fallbackConfig,
				Secret: types.StringValue(""),
			}

			var diags diag.Diagnostics
			model := alertChannelToModel(ctx, apiChannel, fallback, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %s", diags.Errors())
			}

			var resultConfig map[string]string
			model.Config.ElementsAs(ctx, &resultConfig, false)

			if got := resultConfig["url"]; got != tt.wantURL {
				t.Errorf("config[url] = %q, want %q", got, tt.wantURL)
			}
		})
	}
}

func TestIsRedacted(t *testing.T) {
	tests := []struct {
		name       string
		apiValue   string
		stateValue string
		want       bool
	}{
		{"empty api value", "", "https://hooks.example.com", true},
		{"ellipsis in api value", "https://h...m/alert", "https://hooks.example.com/alert", true},
		{"shorter api value", "https://ho*****", "https://hooks.example.com/alert", true},
		{"identical values", "https://hooks.example.com", "https://hooks.example.com", false},
		{"api value longer", "https://hooks.example.com/v2", "https://hooks.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRedacted(tt.apiValue, tt.stateValue); got != tt.want {
				t.Errorf("isRedacted(%q, %q) = %v, want %v", tt.apiValue, tt.stateValue, got, tt.want)
			}
		})
	}
}
