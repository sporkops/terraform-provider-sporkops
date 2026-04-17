package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sporkops/spork-go"
)

var _ datasource.DataSource = &MaintenanceWindowsDataSource{}
var _ datasource.DataSourceWithConfigure = &MaintenanceWindowsDataSource{}

// MaintenanceWindowsDataSource lists maintenance windows with optional
// state filtering. Mirrors the alert-channels plural data source.
type MaintenanceWindowsDataSource struct {
	client *spork.Client
}

type MaintenanceWindowsDataSourceModel struct {
	State              types.String                       `tfsdk:"state"`
	MaintenanceWindows []MaintenanceWindowDataSourceModel `tfsdk:"maintenance_windows"`
}

func NewMaintenanceWindowsDataSource() datasource.DataSource {
	return &MaintenanceWindowsDataSource{}
}

func (d *MaintenanceWindowsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_windows"
}

func (d *MaintenanceWindowsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches all Spork maintenance windows with optional state filtering.",
		MarkdownDescription: "Fetches all [Spork](https://sporkops.com) maintenance windows. Filter by state with the optional `state` attribute.",
		Attributes: map[string]schema.Attribute{
			"state": schema.StringAttribute{
				Optional:            true,
				Description:         "Filter by lifecycle state.",
				MarkdownDescription: "Filter by lifecycle state. One of: `scheduled`, `in_progress`, `completed`, `cancelled`.",
			},
			"maintenance_windows": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of maintenance windows matching the filter.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                  schema.StringAttribute{Computed: true},
						"name":                schema.StringAttribute{Computed: true},
						"description":         schema.StringAttribute{Computed: true},
						"monitor_ids":         schema.SetAttribute{Computed: true, ElementType: types.StringType},
						"tag_selectors":       schema.SetAttribute{Computed: true, ElementType: types.StringType},
						"all_monitors":        schema.BoolAttribute{Computed: true},
						"timezone":            schema.StringAttribute{Computed: true},
						"start_at":            schema.StringAttribute{Computed: true},
						"end_at":              schema.StringAttribute{Computed: true},
						"recurrence_type":     schema.StringAttribute{Computed: true},
						"recurrence_days":     schema.ListAttribute{Computed: true, ElementType: types.Int64Type},
						"recurrence_until":    schema.StringAttribute{Computed: true},
						"suppress_alerts":     schema.BoolAttribute{Computed: true},
						"exclude_from_uptime": schema.BoolAttribute{Computed: true},
						"pause_checks":        schema.BoolAttribute{Computed: true},
						"state":               schema.StringAttribute{Computed: true},
						"created_at":          schema.StringAttribute{Computed: true},
						"updated_at":          schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *MaintenanceWindowsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*spork.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *spork.Client, got something else. Please report this issue to the provider developers.",
		)
		return
	}
	d.client = client
}

func (d *MaintenanceWindowsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MaintenanceWindowsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	windows, err := d.client.ListMaintenanceWindows(ctx)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Error listing maintenance windows", err)
		return
	}

	filterState := ""
	if !config.State.IsNull() && !config.State.IsUnknown() {
		filterState = config.State.ValueString()
	}

	results := make([]MaintenanceWindowDataSourceModel, 0, len(windows))
	for _, w := range windows {
		if filterState != "" && w.State != filterState {
			continue
		}
		results = append(results, maintenanceWindowDataModelFromAPI(ctx, w, &resp.Diagnostics))
	}

	state := MaintenanceWindowsDataSourceModel{
		State:              config.State,
		MaintenanceWindows: results,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
