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
var _ resource.Resource = (*droneResource)(nil)
var _ resource.ResourceWithConfigure = (*droneResource)(nil)

func NewDroneResource() resource.Resource { return &droneResource{} }

type droneResource struct {
	cl *client.ClientWithResponses
}

func (r *droneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *droneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_drone"
}

func (r *droneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.DroneResourceSchema()
}

func (r *droneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data schema.DroneModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiDrone, err := data.ToModelJSON(ctx)
	if err != nil {
		resp.Diagnostics.AddError("unable to create drone", "unable to convert drone data: "+err.Error())
		return
	}

	apiResp, err := r.cl.PostDronesCreateWithResponse(ctx, apiDrone)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create drone: %s", err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to create drone: %s", apiResp.Status()))
		return
	}

	if apiResp.JSON200.Id == nil {
		resp.Diagnostics.AddError("res id in response cannon be ni;", fmt.Sprintf("bad response data: %v", apiResp.JSON200))
		return
	}
	id := *apiResp.JSON200.Id

	getResp, err := r.cl.GetDronesGetWithResponse(ctx, &client.GetDronesGetParams{Id: id})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created drone '%s': %s", data.ID.String(), err))
		return
	}
	if getResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", getResp.StatusCode()), fmt.Sprintf("Unable to read created drone '%s': %s", data.ID.String(), getResp.Status()))
		return
	}
	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API when reading created drone '%s'", data.ID.String()))
		return
	}
	if err := data.FillFromResp(ctx, *getResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *droneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data schema.DroneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.cl.GetDronesGetWithResponse(ctx, &client.GetDronesGetParams{Id: data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get drone '%s': %s", data.ID.String(), err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to get drone '%s': %s", data.ID.String(), apiResp.Status()))
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API for drone '%s'", data.ID.String()))
		return
	}

	if err := data.FillFromResp(ctx, *apiResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *droneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data schema.DroneModel
	var stateData schema.DroneModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiDrone, err := data.ToModelJSON(ctx)
	if err != nil {
		resp.Diagnostics.AddError("unable to update drone", "unable to convert drone data: "+err.Error())
		return
	}
	id := stateData.ID.ValueString()
	apiDrone.Id = &id

	apiResp, err := r.cl.PostDronesUpdateWithResponse(ctx, apiDrone)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update drone '%s': %s", id, err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to update drone '%s': %s", id, apiResp.Status()))
		return
	}

	getResp, err := r.cl.GetDronesGetWithResponse(ctx, &client.GetDronesGetParams{Id: id})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated drone '%s': %s", id, err))
		return
	}
	if getResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", getResp.StatusCode()), fmt.Sprintf("Unable to read updated drone '%s': %s", id, getResp.Status()))
		return
	}
	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API when reading updated drone '%s'", id))
		return
	}
	if err := data.FillFromResp(ctx, *getResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *droneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data schema.DroneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() || data.ID.IsUnknown() {
		resp.Diagnostics.AddError("drone id cannot be empty", "drone id cannot be null")
		return
	}

	apiResp, err := r.cl.DeleteDronesDeleteWithResponse(ctx, &client.DeleteDronesDeleteParams{Id: data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete drone '%s': %s", data.ID.String(), err))
		return
	}

	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete drone '%s': %s", data.ID.String(), apiResp.Status()))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Drone '%s' deleted successfully", data.ID.String()))
}
