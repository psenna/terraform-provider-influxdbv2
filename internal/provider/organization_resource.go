// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &organizationResource{}
var _ resource.ResourceWithImportState = &organizationResource{}

func OrganizationResource() resource.Resource {
	return &organizationResource{}
}

// organizationResource defines the resource implementation.
type organizationResource struct {
	client influxdb2.Client
}

// organizationResourceModel describes the resource data model.
type organizationResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Id          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *organizationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *organizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "organization resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Organization name",
				Required:            true,
				Optional:            false,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Organizatin description",
				Required:            false,
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Organizatin id",
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

func (r *organizationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state organizationResourceModel

	// Read Terraform plan state into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var organization domain.Organization
	organization.Name = state.Name.ValueString()
	organization.Description = state.Description.ValueStringPointer()

	newOrganization, err := r.client.OrganizationsAPI().CreateOrganization(context.Background(), &organization)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating organization",
			fmt.Sprintf("Error: %s", err),
		)

		return
	}

	state.Id = types.StringPointerValue(newOrganization.Id)

	state.Name = types.StringValue(newOrganization.Name)

	state.Description = types.StringPointerValue(newOrganization.Description)

	state.Status = types.StringPointerValue((*string)(newOrganization.Status))

	state.CreatedAt = types.StringValue(newOrganization.CreatedAt.String())

	state.UpdatedAt = types.StringValue(newOrganization.UpdatedAt.String())

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	organization, err := r.client.OrganizationsAPI().FindOrganizationByID(context.Background(), state.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading organization",
			fmt.Sprintf("Could not read organization %s with ID %s : %s", state.Name, state.Id, err),
		)

		return
	}

	state.Id = types.StringPointerValue(organization.Id)

	state.Description = types.StringPointerValue(organization.Description)

	state.Status = types.StringPointerValue((*string)(organization.Status))

	state.CreatedAt = types.StringValue(organization.CreatedAt.String())

	state.UpdatedAt = types.StringValue(organization.UpdatedAt.String())

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state organizationResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	organization, err := r.client.OrganizationsAPI().FindOrganizationByID(context.Background(), state.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading organization",
			fmt.Sprintf("Could not read organization %s with ID %s : %s", plan.Name, state.Id.ValueString(), err),
		)

		return
	}

	if organization.Id == nil {
		resp.Diagnostics.AddError(
			"Error reading organization",
			fmt.Sprintf("Could not read organization %s with ID %s : Organization not found", plan.Name, state.Id.ValueString()),
		)

		return
	}

	organization.Name = plan.Name.ValueString()
	organization.Description = plan.Description.ValueStringPointer()

	organization, err = r.client.OrganizationsAPI().UpdateOrganization(context.Background(), organization)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating organization",
			fmt.Sprintf("Could not update organization %s with ID %s : %s", plan.Name, plan.Id, err),
		)

		return
	}

	plan.Id = types.StringPointerValue(organization.Id)

	plan.Description = types.StringPointerValue(organization.Description)

	plan.Status = types.StringPointerValue((*string)(organization.Status))

	plan.CreatedAt = types.StringValue(organization.CreatedAt.String())

	plan.UpdatedAt = types.StringValue(organization.UpdatedAt.String())

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.OrganizationsAPI().DeleteOrganizationWithID(context.Background(), state.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleteing organization",
			fmt.Sprintf("Could not update organization %s with ID %s : %s", state.Name, state.Id, err),
		)

		return
	}

}

func (r *organizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
