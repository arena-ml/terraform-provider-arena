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
var _ resource.Resource = (*stepResource)(nil)
var _ resource.ResourceWithConfigure = (*stepResource)(nil)

func NewStepResource() resource.Resource { return &stepResource{} }

type stepResource struct {
	cl *client.ClientWithResponses
}

func (r *stepResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *stepResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_step"
}

func (r *stepResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Resource schema: mirrors NodeStep model. Some attributes are computed from server.
	resp.Schema = schema.NodeStepResourceSchema()
}

func (r *stepResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data schema.NodeStep

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.upsertResource(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("error saving step", "Failed to create step "+err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, created)...)
}

func (r *stepResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data schema.NodeStep
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.cl.GetPipelineNodesOneWithResponse(ctx, &client.GetPipelineNodesOneParams{Id: data.ID.ValueString(), Kind: "step"})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get step '%s': %s", data.ID.String(), err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to get step '%s': %s", data.ID.String(), apiResp.Status()))
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API for step '%s'", data.ID.String()))
		return
	}

	steps := apiResp.JSON200.Steps
	if steps == nil || len(*steps) != 1 {
		resp.Diagnostics.AddError("unexpected response body", "response should have exactly one step node")
		return
	}

	stepNode := (*steps)[0]

	if err := data.FillFromResp(ctx, stepNode); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *stepResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data schema.NodeStep
	var stateData schema.NodeStep

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = stateData.ID

	created, err := r.upsertResource(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("error saving step", "Failed to create step "+err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, created)...)
}

func (r *stepResource) upsertResource(ctx context.Context, data *schema.NodeStep) (*schema.NodeStep, error) {
	sNode, err := data.ToModelJSON(ctx)
	if err != nil {
		return nil, err
	}

	steps := []client.EntStep{sNode}
	payload := client.ModelPipelineNodes{
		PipelineId: data.PipelineID.ValueStringPointer(),
		Inputs:     nil,
		Outputs:    nil,
		Steps:      &steps,
	}

	// Call Create
	apiResp, err := r.cl.PostPipelineNodesCreateWithResponse(ctx, payload)
	if err != nil {
		return nil, err
	}
	if apiResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("api call failed : %d", apiResp.StatusCode())
	}

	// Parse response to get ID
	if apiResp.JSON200 == nil || apiResp.JSON200.Steps == nil || len(*apiResp.JSON200.Steps) != 1 {
		return nil, fmt.Errorf("invalid api response , step len: \n\n%s\n", pretty.Sprint(*apiResp.JSON200))
	}

	created := (*apiResp.JSON200.Steps)[0]

	if created.Id == nil {
		return nil, fmt.Errorf("invalid resource id in api response")
	}

	newData := &schema.NodeStep{}
	newData.PipelineID = types.StringValue(*apiResp.JSON200.PipelineId)

	// Fill TF model from response
	err = newData.FillFromResp(ctx, created)
	return newData, err
}

func (r *stepResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data schema.NodeStep
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodeIds := []string{data.ID.String()}

	payload := client.ModelPipelineNodeIDs{
		PipelineId: data.PipelineID.ValueStringPointer(),
		Stepids:    &nodeIds,
	}

	apiResp, err := r.cl.DeletePipelineNodesDeleteWithResponse(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete step '%s': %s", data.ID.String(), err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to delete step '%s': %s", data.ID.String(), apiResp.Status()))
		return
	}
}
