package provider

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sporkops/spork-go"
)

const defaultAPIBaseURL = "https://api.sporkops.com/v1"

var _ provider.Provider = &SporkProvider{}

type SporkProvider struct {
	version string
}

type SporkProviderModel struct {
	APIKey         types.String `tfsdk:"api_key"`
	OrganizationID types.String `tfsdk:"organization_id"`
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
			"organization_id": schema.StringAttribute{
				Optional: true,
				Description: "Organization ID for org-scoped resources (monitors, alert channels, " +
					"members, maintenance windows). Can also be set via the SPORK_ORG_ID " +
					"environment variable. When omitted, the provider auto-resolves the " +
					"organization by listing memberships for the API key — which works " +
					"transparently for keys bound to a single organization. Set explicitly " +
					"when the caller belongs to multiple organizations.",
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
		baseURL = defaultAPIBaseURL
	}

	if err := validateAPIBaseURL(baseURL); err != nil {
		resp.Diagnostics.AddError("Invalid SPORK_API_URL", err.Error())
		return
	}

	orgID := strings.TrimSpace(os.Getenv("SPORK_ORG_ID"))
	if !config.OrganizationID.IsNull() && !config.OrganizationID.IsUnknown() {
		orgID = strings.TrimSpace(config.OrganizationID.ValueString())
	}

	clientOpts := []spork.Option{
		spork.WithAPIKey(apiKey),
		spork.WithBaseURL(baseURL),
		spork.WithUserAgent("spork-terraform/" + p.version),
	}
	if orgID != "" {
		clientOpts = append(clientOpts, spork.WithOrganization(orgID))
	}

	client := spork.NewClient(clientOpts...)

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
		NewMaintenanceWindowDataSource,
		NewMaintenanceWindowsDataSource,
	}
}

// validateAPIBaseURL rejects SPORK_API_URL values that would leak the API key
// to a plaintext or internal-network endpoint. The default production URL is
// always allowed; any override must parse cleanly, use https, and resolve to a
// non-loopback, non-private, non-link-local host.
func validateAPIBaseURL(raw string) error {
	if raw == defaultAPIBaseURL {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("SPORK_API_URL is not a valid URL: %w", err)
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return fmt.Errorf("SPORK_API_URL must use https (got %q); the API key is sent with every request and must not transit plaintext", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("SPORK_API_URL must include a hostname")
	}
	// Block literal IPs in loopback, private, link-local, or unspecified
	// ranges. Hostnames that resolve dynamically are trusted (the server's
	// TLS cert remains the authentication anchor); guarding those here would
	// race DNS and give a false sense of security.
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsMulticast() {
			return fmt.Errorf("SPORK_API_URL host %q is in a loopback or private range; API keys must not be sent to internal endpoints", host)
		}
	}
	// Catch well-known hostnames that map to loopback or cloud metadata.
	lower := strings.ToLower(host)
	switch lower {
	case "localhost", "ip6-localhost", "ip6-loopback", "metadata", "metadata.google.internal":
		return fmt.Errorf("SPORK_API_URL host %q is not permitted; API keys must not be sent to internal endpoints", host)
	}
	return nil
}
