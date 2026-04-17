package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sporkops/spork-go"
)

var _ provider.Provider = &SporkProvider{}

type SporkProvider struct {
	version string
}

type SporkProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SporkProvider{
			version: version,
		}
	}
}

func (p *SporkProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "spork"
	resp.Version = p.version
}

func (p *SporkProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Spork provider is used to manage uptime monitors, alert channels, maintenance windows, and status pages via the Spork API.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Spork API key. Can also be set via the SPORK_API_KEY environment variable.",
			},
		},
	}
}

func (p *SporkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config SporkProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the API key is unknown (e.g., computed from another resource),
	// we cannot configure the client yet. Return early to allow planning.
	if config.APIKey.IsUnknown() {
		return
	}

	apiKey := os.Getenv("SPORK_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Spork API key required",
			"Set your API key to authenticate with Spork:\n\n"+
				"  export SPORK_API_KEY=\"your-api-key\"\n\n"+
				"Don't have an account? Sign up free:\n"+
				"  https://sporkops.com/signup?ref=terraform\n\n"+
				"Generate an API key:\n"+
				"  1. Install the CLI: brew install sporkops/tap/spork\n"+
				"  2. Run: spork login\n"+
				"  3. Run: spork api-key create\n"+
				"  4. Export the key: export SPORK_API_KEY=\"sk_...\"\n\n"+
				"Docs: https://sporkops.com/docs",
		)
		return
	}

	baseURL := os.Getenv("SPORK_API_URL")
	if baseURL == "" {
		baseURL = "https://api.sporkops.com/v1"
	}

	if !strings.HasPrefix(baseURL, "https://") {
		resp.Diagnostics.AddWarning(
			"Insecure API URL",
			"SPORK_API_URL does not use HTTPS. Your API key will be sent in cleartext. "+
				"This is acceptable for local development but should not be used in production.",
		)
	}

	client := spork.NewClient(
		spork.WithAPIKey(apiKey),
		spork.WithBaseURL(baseURL),
		spork.WithUserAgent("spork-terraform/"+p.version),
	)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *SporkProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMonitorResource,
		NewAlertChannelResource,
		NewStatusPageResource,
		NewMemberResource,
		NewMaintenanceWindowResource,
	}
}

func (p *SporkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewMonitorDataSource,
		NewAlertChannelDataSource,
		NewMonitorsDataSource,
		NewAlertChannelsDataSource,
		NewStatusPageDataSource,
		NewStatusPagesDataSource,
		NewMembersDataSource,
		NewOrganizationDataSource,
	}
}
