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
var _ resource.Resource = &bucketResource{}
var _ resource.ResourceWithImportState = &bucketResource{}

func BucketResource() resource.Resource {
	return &bucketResource{}
}

// bucketResource defines the resource implementation.
type bucketResource struct {
	client influxdb2.Client
}

// bucketResourceModel describes the resource data model.
type bucketResourceModel struct {
	Name          types.String                `tfsdk:"name"`
	Id            types.String                `tfsdk:"id"`
	OrgID         types.String                `tfsdk:"org_id"`
	Description   types.String                `tfsdk:"description"`
	RetentioRules []bucketRetentionRulesModel `tfsdk:"retention_rules"`
	RP            types.String                `tfsdk:"rp"`
	ScehmaType    types.String                `tfsdk:"schema_type"`
	CreatedAt     types.String                `tfsdk:"created_at"`
	UpdatedAt     types.String                `tfsdk:"updated_at"`
}

type bucketRetentionRulesModel struct {
	EverySeconds  types.Int64  `tfsdk:"every_seconds"`
	RetentionType types.String `tfsdk:"retention_type"`
}

func (r *bucketResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}

func (r *bucketResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "bucket resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Bucket name",
				Required:            true,
				Optional:            false,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Bucket id",
				Computed:            true,
			},
			"org_id": schema.StringAttribute{
				MarkdownDescription: "Bucket id",
				Required:            true,
				Optional:            false,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Bucket description",
				Required:            false,
				Optional:            true,
			},
			"retention_rules": schema.ListNestedAttribute{
				Required:            true,
				Optional:            false,
				MarkdownDescription: "Bucket retention rules",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"every_seconds": schema.Int64Attribute{
							Required: true,
						},
						"retention_type": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"rp": schema.StringAttribute{
				MarkdownDescription: "Bucket retention policy",
				Required:            false,
				Optional:            true,
			},
			"schema_type": schema.StringAttribute{
				MarkdownDescription: "Bucket retention policy",
				Required:            false,
				Optional:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Bucket creation date",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Bucket update date",
				Computed:            true,
			},
		},
	}
}

func (r *bucketResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state bucketResourceModel

	// Read Terraform plan state into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var retentionRules []domain.RetentionRule

	for _, rule := range state.RetentioRules {
		retentionRules = append(retentionRules, domain.RetentionRule{
			EverySeconds: rule.EverySeconds.ValueInt64(),
			Type:         (*domain.RetentionRuleType)(rule.RetentionType.ValueStringPointer()),
		})
	}

	var bucket domain.Bucket
	bucket.Name = state.Name.ValueString()
	bucket.Description = state.Description.ValueStringPointer()
	bucket.OrgID = state.OrgID.ValueStringPointer()
	bucket.Rp = state.RP.ValueStringPointer()
	bucket.RetentionRules = retentionRules

	newBucket, err := r.client.BucketsAPI().CreateBucket(context.Background(), &bucket)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bucket",
			fmt.Sprintf("Error: %s", err),
		)

		return
	}

	var newRetentionRules []bucketRetentionRulesModel

	for _, rule := range newBucket.RetentionRules {
		newRetentionRules = append(newRetentionRules, bucketRetentionRulesModel{
			EverySeconds:  types.Int64Value(rule.EverySeconds),
			RetentionType: types.StringPointerValue((*string)(rule.Type)),
		})
	}

	state.Id = types.StringPointerValue(newBucket.Id)
	state.Name = types.StringValue(newBucket.Name)
	state.Description = types.StringPointerValue(newBucket.Description)
	state.OrgID = types.StringPointerValue(newBucket.OrgID)
	state.RP = types.StringPointerValue(newBucket.Rp)
	state.RetentioRules = newRetentionRules
	state.ScehmaType = types.StringPointerValue((*string)(newBucket.SchemaType))
	state.CreatedAt = types.StringValue(newBucket.CreatedAt.String())
	state.UpdatedAt = types.StringValue(newBucket.UpdatedAt.String())

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucketResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := r.client.BucketsAPI().FindBucketByID(context.Background(), state.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bucket",
			fmt.Sprintf("Could not read bucket %s with ID %s : %s", state.Name, state.Id, err),
		)

		return
	}

	var actualRetentionRules []bucketRetentionRulesModel

	for _, rule := range bucket.RetentionRules {
		actualRetentionRules = append(actualRetentionRules, bucketRetentionRulesModel{
			EverySeconds:  types.Int64Value(rule.EverySeconds),
			RetentionType: types.StringPointerValue((*string)(rule.Type)),
		})
	}

	state.Id = types.StringPointerValue(bucket.Id)
	state.Name = types.StringValue(bucket.Name)
	state.Description = types.StringPointerValue(bucket.Description)
	state.OrgID = types.StringPointerValue(bucket.OrgID)
	state.RP = types.StringPointerValue(bucket.Rp)
	state.RetentioRules = actualRetentionRules
	state.ScehmaType = types.StringPointerValue((*string)(bucket.SchemaType))
	state.CreatedAt = types.StringValue(bucket.CreatedAt.String())
	state.UpdatedAt = types.StringValue(bucket.UpdatedAt.String())
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bucketResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state bucketResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := r.client.BucketsAPI().FindBucketByID(context.Background(), state.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bucket",
			fmt.Sprintf("Could not read bucket %s with ID %s : %s", plan.Name, state.Id.ValueString(), err),
		)

		return
	}

	if bucket.Id == nil {
		resp.Diagnostics.AddError(
			"Error reading bucket",
			fmt.Sprintf("Could not read bucket %s with ID %s : Bucket not found", plan.Name, state.Id.ValueString()),
		)

		return
	}

	var retentionRules []domain.RetentionRule

	for _, rule := range state.RetentioRules {
		retentionRules = append(retentionRules, domain.RetentionRule{
			EverySeconds: rule.EverySeconds.ValueInt64(),
			Type:         (*domain.RetentionRuleType)(rule.RetentionType.ValueStringPointer()),
		})
	}

	bucket.Name = plan.Name.ValueString()
	bucket.Description = plan.Description.ValueStringPointer()
	bucket.RetentionRules = retentionRules

	bucket, err = r.client.BucketsAPI().UpdateBucket(context.Background(), bucket)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating bucket",
			fmt.Sprintf("Could not update bucket %s with ID %s : %s", plan.Name, plan.Id, err),
		)

		return
	}

	plan.Id = types.StringPointerValue(bucket.Id)

	plan.Description = types.StringPointerValue(bucket.Description)

	plan.CreatedAt = types.StringValue(bucket.CreatedAt.String())

	plan.UpdatedAt = types.StringValue(bucket.UpdatedAt.String())

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bucketResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.BucketsAPI().DeleteBucketWithID(context.Background(), state.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleteing bucket",
			fmt.Sprintf("Could not update bucket %s with ID %s : %s", state.Name, state.Id, err),
		)

		return
	}

}

func (r *bucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
