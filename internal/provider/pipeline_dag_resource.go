// Copyright (c) ArenaML Labs Pvt Ltd.

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/arena-ml/terraform-provider-arenaml/generator/client"
	"github.com/arena-ml/terraform-provider-arenaml/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kr/pretty"
)

// Ensure implementation satisfies interfaces
var _ resource.Resource = (*pipelineDagResource)(nil)
var _ resource.ResourceWithConfigure = (*pipelineDagResource)(nil)

func NewPipelineDagResource() resource.Resource { return &pipelineDagResource{} }

type pipelineDagResource struct {
	cl *client.ClientWithResponses
}

func (r *pipelineDagResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *pipelineDagResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_dag"
}

func (r *pipelineDagResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Resource schema: mirrors PipelineDag model
	resp.Schema = schema.PipelineDagResourceSchema()
}

func (r *pipelineDagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data schema.PipelineDag

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.upsertResource(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("error saving pipeline DAG", "Failed to create pipeline DAG "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, created)...)
}

func (r *pipelineDagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data schema.PipelineDag
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.PipelineID.ValueString()
	apiResp, err := r.cl.GetPipelineDagWithResponse(ctx, &client.GetPipelineDagParams{
		Id: &id,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get pipeline DAG for pipeline '%s': %s", data.PipelineID.String(), err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to get pipeline DAG for pipeline '%s': %s", data.PipelineID.String(), apiResp.Status()))
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API for pipeline DAG '%s'", data.PipelineID.String()))
		return
	}

	if err := data.FillFromResp(ctx, *apiResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *pipelineDagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data schema.PipelineDag
	var stateData schema.PipelineDag

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.PipelineID = stateData.PipelineID

	created, err := r.upsertResource(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("error saving pipeline DAG", "Failed to update pipeline DAG "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, created)...)
}

func (r *pipelineDagResource) upsertResource(ctx context.Context, data *schema.PipelineDag) (*schema.PipelineDag, error) {
	dagModel, err := data.ToModelJSON(ctx)
	if err != nil {
		return nil, err
	}

	// Call Create or Update (they use the same payload structure)
	apiResp, err := r.cl.PostPipelineDagCreateWithResponse(ctx, dagModel)
	if err != nil {
		return nil, err
	}
	if apiResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("api call failed : %d", apiResp.StatusCode())
	}

	// Parse response
	if apiResp.JSON200 == nil {
		return nil, fmt.Errorf("invalid api response: %s", pretty.Sprint(apiResp))
	}

	newData := &schema.PipelineDag{}
	newData.PipelineID = types.StringValue(*apiResp.JSON200.PipelineId)

	// Fill TF model from response
	err = newData.FillFromResp(ctx, *apiResp.JSON200)
	return newData, err
}

func (r *pipelineDagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data schema.PipelineDag
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dagModel, err := data.ToModelJSON(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert pipeline DAG '%s' for deletion: %s", data.PipelineID.String(), err))
		return
	}

	apiResp, err := r.cl.DeletePipelineDagDeleteWithResponse(ctx, dagModel)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete pipeline DAG '%s': %s", data.PipelineID.String(), err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to delete pipeline DAG '%s': %s", data.PipelineID.String(), apiResp.Status()))
		return
	}
}
