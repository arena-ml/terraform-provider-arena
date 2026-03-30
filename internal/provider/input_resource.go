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
var _ resource.Resource = (*inputResource)(nil)
var _ resource.ResourceWithConfigure = (*inputResource)(nil)

func NewInputResource() resource.Resource { return &inputResource{} }

type inputResource struct {
	cl *client.ClientWithResponses
}

func (r *inputResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Use the same underlying model as the data source

func (r *inputResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_input"
}

func (r *inputResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Resource schema: mirrors InputNode model. Some attributes are computed from server.
	resp.Schema = schema.InputNodeResourceSchema()
}

func (r *inputResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data schema.InputNode

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.upsertResource(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("error saving input", "Failed to create input "+err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, created)...)
}

func (r *inputResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data schema.InputNode
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.cl.GetPipelineNodesOneWithResponse(ctx, &client.GetPipelineNodesOneParams{Id: data.ID.ValueString(), Kind: "input"})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get input '%s': %s", data.ID.String(), err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to get input '%s': %s", data.ID.String(), apiResp.Status()))
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API for input '%s'", data.ID.String()))
		return
	}

	inputs := apiResp.JSON200.Inputs
	if inputs == nil || len(*inputs) != 1 {
		resp.Diagnostics.AddError("unexpected response body", "response should have exactly one input node")
		return
	}

	inputNode := (*inputs)[0]

	if err := data.FillFromResp(ctx, inputNode); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *inputResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data schema.InputNode
	var stateData schema.InputNode

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = stateData.ID

	created, err := r.upsertResource(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("error saving input", "Failed to create input "+err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, created)...)
}

func (r *inputResource) upsertResource(ctx context.Context, data *schema.InputNode) (*schema.InputNode, error) {
	iNode, err := data.ToModelJSON(ctx)
	if err != nil {
		return nil, err
	}

	inputs := []client.EntInput{iNode}
	payload := client.ModelPipelineNodes{
		PipelineId: data.PipelineID.ValueStringPointer(),
		Inputs:     &inputs,
		Outputs:    nil,
		Steps:      nil,
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
	if apiResp.JSON200 == nil || apiResp.JSON200.Inputs == nil || len(*apiResp.JSON200.Inputs) != 1 {
		return nil, fmt.Errorf("invalid api response , input len: \n\n%s\n", pretty.Sprint(*apiResp.JSON200))
	}

	created := (*apiResp.JSON200.Inputs)[0]

	if created.Id == nil {
		return nil, fmt.Errorf("invalid resource id in api response")
	}

	newData := &schema.InputNode{}
	newData.PipelineID = types.StringValue(*apiResp.JSON200.PipelineId)

	// Fill TF model from response
	err = newData.FillFromResp(ctx, created)
	return newData, err
}

func (r *inputResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data schema.InputNode
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodeIds := []string{data.ID.String()}

	payload := client.ModelPipelineNodeIDs{
		PipelineId: data.PipelineID.ValueStringPointer(),
		Inputs:     &nodeIds,
	}

	apiResp, err := r.cl.DeletePipelineNodesDeleteWithResponse(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete input '%s': %s", data.ID.String(), err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to delete input '%s': %s", data.ID.String(), apiResp.Status()))
		return
	}
}
