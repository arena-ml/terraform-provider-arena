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
var _ resource.Resource = (*orgResource)(nil)
var _ resource.ResourceWithConfigure = (*orgResource)(nil)

func NewOrgResource() resource.Resource { return &orgResource{} }

type orgResource struct {
	cl *client.ClientWithResponses
}

func (r *orgResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *orgResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

func (r *orgResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.OrgResourceSchema()
}

func (r *orgResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data schema.Org

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert TF model to API model
	apiOrg, err := data.ToModelJSON(ctx)
	if err != nil {
		resp.Diagnostics.AddError("unable to create org", "unable to convert org data: "+err.Error())
		return
	}

	// Call Create API
	apiResp, err := r.cl.PostIamOrgCreateWithResponse(ctx, apiOrg)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create org: %s", err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to create org: %s", apiResp.Status()))
		return
	}

	// Read back the created org
	getResp, err := r.cl.GetIamOrgGetWithResponse(ctx, &client.GetIamOrgGetParams{Id: data.ID.ValueStringPointer()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created org '%s': %s", data.ID.String(), err))
		return
	}
	if getResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", getResp.StatusCode()), fmt.Sprintf("Unable to read created org '%s': %s", data.ID.String(), getResp.Status()))
		return
	}
	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API when reading created org '%s'", data.ID.String()))
		return
	}
	if err := data.FillFromResp(ctx, *getResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *orgResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data schema.Org
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.cl.GetIamOrgGetWithResponse(ctx, &client.GetIamOrgGetParams{Id: data.ID.ValueStringPointer()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get org '%s': %s", data.ID.String(), err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to get org '%s': %s", data.ID.String(), apiResp.Status()))
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API for org '%s'", data.ID.String()))
		return
	}

	if err := data.FillFromResp(ctx, *apiResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *orgResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data schema.Org
	var stateData schema.Org

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert TF model to API model
	apiOrg, err := data.ToModelJSON(ctx)
	if err != nil {
		resp.Diagnostics.AddError("unable to update org", "unable to convert org data: "+err.Error())
		return
	}
	id := stateData.ID.ValueString()
	apiOrg.Id = &id

	// Call Update API
	apiResp, err := r.cl.PostIamOrgUpdateWithResponse(ctx, apiOrg)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update org '%s': %s", id, err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to update org '%s': %s", id, apiResp.Status()))
		return
	}

	// Read back the updated org
	getResp, err := r.cl.GetIamOrgGetWithResponse(ctx, &client.GetIamOrgGetParams{Id: &id})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated org '%s': %s", id, err))
		return
	}
	if getResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", getResp.StatusCode()), fmt.Sprintf("Unable to read updated org '%s': %s", id, getResp.Status()))
		return
	}
	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API when reading updated org '%s'", id))
		return
	}
	if err := data.FillFromResp(ctx, *getResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *orgResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data schema.Org
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() || data.ID.IsUnknown() {
		resp.Diagnostics.AddError("org id cannot be empty", "org id cannot be empty")
		return
	}

	apiResp, err := r.cl.DeleteIamOrgDeleteWithResponse(ctx, &client.DeleteIamOrgDeleteParams{Id: data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete org '%s': %s", data.ID.String(), err))
		return
	}

	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete org '%s': %s", data.ID.String(), apiResp.Status()))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Org '%s' removed from Terraform state.", data.ID.String()))
}
