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
var _ resource.Resource = (*sensorResource)(nil)
var _ resource.ResourceWithConfigure = (*sensorResource)(nil)

func NewSensorResource() resource.Resource { return &sensorResource{} }

type sensorResource struct {
	cl *client.ClientWithResponses
}

func (r *sensorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sensorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sensor"
}

func (r *sensorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.SensorResourceSchema()
}

func (r *sensorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data schema.SensorModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiSensor, err := data.ToModelJSON(ctx)
	if err != nil {
		resp.Diagnostics.AddError("unable to create sensor", "unable to convert sensor data: "+err.Error())
		return
	}

	apiResp, err := r.cl.PostSensorsCreateWithResponse(ctx, apiSensor)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create sensor: %s", err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to create sensor: %s", apiResp.Status()))
		return
	}

	if apiResp.JSON200.Id == nil {
		resp.Diagnostics.AddError("res id in response cannon be ni;", fmt.Sprintf("bad response data: %v", apiResp.JSON200))
		return
	}
	id := *apiResp.JSON200.Id

	getResp, err := r.cl.GetSensorsGetWithResponse(ctx, &client.GetSensorsGetParams{Id: id})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created sensor '%s': %s", data.ID.String(), err))
		return
	}
	if getResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", getResp.StatusCode()), fmt.Sprintf("Unable to read created sensor '%s': %s", data.ID.String(), getResp.Status()))
		return
	}
	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API when reading created sensor '%s'", data.ID.String()))
		return
	}
	if err := data.FillFromResp(ctx, getResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sensorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data schema.SensorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.cl.GetSensorsGetWithResponse(ctx, &client.GetSensorsGetParams{Id: data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get sensor '%s': %s", data.ID.String(), err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to get sensor '%s': %s", data.ID.String(), apiResp.Status()))
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API for sensor '%s'", data.ID.String()))
		return
	}

	if err := data.FillFromResp(ctx, apiResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sensorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data schema.SensorModel
	var stateData schema.SensorModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiSensor, err := data.ToModelJSON(ctx)
	if err != nil {
		resp.Diagnostics.AddError("unable to update sensor", "unable to convert sensor data: "+err.Error())
		return
	}
	id := stateData.ID.ValueString()
	apiSensor.Id = &id

	apiResp, err := r.cl.PostSensorsUpdateWithResponse(ctx, apiSensor)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update sensor '%s': %s", id, err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to update sensor '%s': %s", id, apiResp.Status()))
		return
	}

	getResp, err := r.cl.GetSensorsGetWithResponse(ctx, &client.GetSensorsGetParams{Id: id})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated sensor '%s': %s", id, err))
		return
	}
	if getResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", getResp.StatusCode()), fmt.Sprintf("Unable to read updated sensor '%s': %s", id, getResp.Status()))
		return
	}
	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API when reading updated sensor '%s'", id))
		return
	}
	if err := data.FillFromResp(ctx, getResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sensorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data schema.SensorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() || data.ID.IsUnknown() {
		resp.Diagnostics.AddError("sensor id cannot be empty", "sensor id cannot be null")
		return
	}

	apiResp, err := r.cl.DeleteSensorsDeleteWithResponse(ctx, &client.DeleteSensorsDeleteParams{Id: data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete sensor '%s': %s", data.ID.String(), err))
		return
	}

	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete sensor '%s': %s", data.ID.String(), apiResp.Status()))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Sensor '%s' deleted successfully", data.ID.String()))
}
