// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// Ensure InfluxdbV2Provider satisfies various provider interfaces.
var _ provider.Provider = &InfluxdbV2Provider{}

// InfluxdbV2Provider defines the provider implementation.
type InfluxdbV2Provider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// InfluxdbV2ProviderModel describes the provider data model.
type InfluxdbV2ProviderModel struct {
	Host   types.String `tfsdk:"host"`
	ApiKey types.String `tfsdk:"api_key"`
}

func (p *InfluxdbV2Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "influxdbv2"
	resp.Version = p.version
}

func (p *InfluxdbV2Provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Influxdb hostname",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Influxdb Api key",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *InfluxdbV2Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config InfluxdbV2ProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown InfluxdbV2 API Host",
			"The provider cannot create the InfluxdbV2 API client as there is an unknown configuration value for the Influxdb API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_HOST environment variable.",
		)
	}

	influxHost := config.Host.ValueString()
	influxCredential := config.ApiKey.ValueString()

	influxClient := influxdb2.NewClient(influxHost, influxCredential)

	resp.DataSourceData = influxClient
	resp.ResourceData = influxClient
}

func (p *InfluxdbV2Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		BucketResource,
		OrganizationResource,
	}
}

func (p *InfluxdbV2Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		OrganizationDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &InfluxdbV2Provider{
			version: version,
		}
	}
}
