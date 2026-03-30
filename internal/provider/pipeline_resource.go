// Copyright (c) ArenaML Labs Pvt Ltd.

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/arena-ml/terraform-provider-arenaml/generator/client"
	"github.com/arena-ml/terraform-provider-arenaml/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure implementation satisfies interfaces
var _ resource.Resource = (*pipelineResource)(nil)
var _ resource.ResourceWithConfigure = (*pipelineResource)(nil)

func NewPipelineResource() resource.Resource { return &pipelineResource{} }

type pipelineResource struct {
	cl *client.ClientWithResponses
}

func (r *pipelineResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cl, ok := req.ProviderData.(*client.ClientWithResponses)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *oapi client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.cl = cl
}

func (r *pipelineResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline"
}

func (r *pipelineResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.PipelineResourceSchema()
}

func (r *pipelineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data schema.Pipeline

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert TF model to API model
	apiPipeline, err := data.ToModelJSON(ctx)
	if err != nil {
		resp.Diagnostics.AddError("unable to create pipeline", "unable to convert pipeline data: "+err.Error())
		return
	}

	// Call Create API
	apiResp, err := r.cl.PostPipelineCreateWithResponse(ctx, apiPipeline)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create pipeline: %s", err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to create pipeline: %s", apiResp.Status()))
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Received nil response from API when creating pipeline")
		return
	}

	// Fill the data from response
	if err := data.FillFromResp(ctx, *apiResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *pipelineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data schema.Pipeline
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.cl.GetPipelineGetWithResponse(ctx, &client.GetPipelineGetParams{Id: data.ID.ValueStringPointer()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get pipeline '%s': %s", data.ID.String(), err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to get pipeline '%s': %s", data.ID.String(), apiResp.Status()))
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API for pipeline '%s'", data.ID.String()))
		return
	}

	if err := data.FillFromResp(ctx, *apiResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *pipelineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data schema.Pipeline
	var stateData schema.Pipeline

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert TF model to API model
	apiPipeline, err := data.ToModelJSON(ctx)
	if err != nil {
		resp.Diagnostics.AddError("unable to update pipeline", "unable to convert pipeline data: "+err.Error())
		return
	}
	id := stateData.ID.ValueString()
	apiPipeline.Id = &id

	// Call Update API
	apiResp, err := r.cl.PostPipelineUpdateWithResponse(ctx, apiPipeline)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update pipeline '%s': %s", id, err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to update pipeline '%s': %s", id, apiResp.Status()))
		return
	}

	// Read back the updated pipeline
	getResp, err := r.cl.GetPipelineGetWithResponse(ctx, &client.GetPipelineGetParams{Id: &id})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated pipeline '%s': %s", id, err))
		return
	}
	if getResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", getResp.StatusCode()), fmt.Sprintf("Unable to read updated pipeline '%s': %s", id, getResp.Status()))
		return
	}
	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API when reading updated pipeline '%s'", id))
		return
	}
	if err := data.FillFromResp(ctx, *getResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *pipelineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data schema.Pipeline
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() || data.ID.IsUnknown() {
		resp.Diagnostics.AddError("pipeline cannnot be empty", "pipeline id can not set")
		return
	}
	clResp, err := r.cl.DeletePipelineDeleteId(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete pipeline '%s': %s", data.ID.String(), err))
		return
	}

	if clResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete pipeline '%s': %s", data.ID.String(), clResp.Status))
		return
	}

	// Note: The API does not provide a delete endpoint for pipelines.
	// This is intentional - pipelines are typically disabled rather than deleted.
	// We just remove it from Terraform state.
	tflog.Info(ctx, fmt.Sprintf("Pipeline '%s' removed from Terraform state. Note: API does not support pipeline deletion.", data.ID.String()))
}
