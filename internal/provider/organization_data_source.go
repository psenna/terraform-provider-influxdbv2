package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &organizationDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationDataSource{}
)

func OrganizationDataSource() datasource.DataSource {
	return &organizationDataSource{}
}

type organizationDataSource struct {
	client influxdb2.Client
}

// OrganizationDataSourceModel describes the data source data model.
type OrganizationDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	Id          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (d *organizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *organizationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Organization data source",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Organization name",
				Required:            true,
				Optional:            false,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Organizatin id",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Organizatin description",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Organizatin description",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Organizatin creation date",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Organizatin update date",
				Computed:            true,
			},
		},
	}
}

func (d *organizationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(influxdb2.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *influxdb2.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client

}

func (d *organizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state OrganizationDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	organization, err := d.client.OrganizationsAPI().FindOrganizationByName(context.Background(), state.Name.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error on request",
			fmt.Sprintf("Error: %s", err),
		)

		return
	}

	state.Id = types.StringPointerValue(organization.Id)

	state.Description = types.StringPointerValue(organization.Description)

	state.Description = types.StringPointerValue((*string)(organization.Status))

	state.CreatedAt = types.StringValue(organization.CreatedAt.String())

	state.UpdatedAt = types.StringValue(organization.UpdatedAt.String())

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
