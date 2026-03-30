// Copyright (c) ArenaML Labs Pvt Ltd.

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/arena-ml/terraform-provider-arenaml/generator/client"
	"github.com/arena-ml/terraform-provider-arenaml/helper"
	internalschema "github.com/arena-ml/terraform-provider-arenaml/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*engineResource)(nil)

func NewEngineResource() resource.Resource {
	return &engineResource{}
}

type engineResource struct {
	cl *client.ClientWithResponses
}

var _ resource.ResourceWithConfigure = (*engineResource)(nil)

func (r *engineResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Use the same model as the data source
type engineResourceModel = internalschema.EngineModel

func (r *engineResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + suffixClusterManager
}

func (r *engineResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = internalschema.EngineResourceSchema(ctx)
}

func (r *engineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data engineResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	engineConfig := client.EnginesEngineConfig{}

	// Set the basic fields
	err := helper.ConvertTfModelToApiJSON(ctx, data, &engineConfig)

	// Handle spec
	if !data.Spec.IsNull() && !data.Spec.IsUnknown() {
		specObj := make(map[string]interface{})
		err := data.Spec.Unmarshal(&specObj)
		if err != nil {
			resp.Diagnostics.AddError("Error unmarshalling spec", fmt.Sprintf("Unable to unmarshal spec: %s", err))
			return
		}
		engineConfig.Spec = &specObj
	}

	// Handle tags
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []internalschema.Tag
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		apiTags := make([]client.TagsTag, 0, len(tags))
		for _, tag := range tags {
			key := tag.Key.ValueString()
			value := tag.Value.ValueString()
			apiTags = append(apiTags, client.TagsTag{
				Key:   &key,
				Value: &value,
			})
		}
		engineConfig.Tags = &apiTags
	}

	// Make the API call
	apiResp, err := r.cl.PostEngineCreateWithResponse(ctx, engineConfig)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create engine: %s", err))
		return
	}

	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Client Error: %d", apiResp.StatusCode()),
			fmt.Sprintf("Unable to create engine: %s", apiResp.Status()),
		)
		return
	}

	// After creating the engine, we need to read it to get the full details
	// First, we need to parse the response to get the ID
	var createdEngine client.EnginesEngineConfig
	err = json.Unmarshal(apiResp.Body, &createdEngine)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing response", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if createdEngine.Id == nil {
		resp.Diagnostics.AddError("Client Error", "Created engine has no ID")
		return
	}

	// Now read the engine using the ID
	getResp, err := r.cl.GetEngineGetWithResponse(ctx, &client.GetEngineGetParams{Id: createdEngine.Id})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created engine: %s", err))
		return
	}

	if getResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Client Error: %d", getResp.StatusCode()),
			fmt.Sprintf("Unable to read created engine: %s", getResp.Status()),
		)
		return
	}

	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Received nil response from API when reading created engine")
		return
	}

	engineResp := *getResp.JSON200

	// Update the model with the response data
	helper.ConvertJSONStructToSimpleTF(ctx, engineResp, &data)

	// Handle tags
	var tagDiag diag.Diagnostics
	var tags []internalschema.Tag

	if engineResp.Tags != nil {
		tags = internalschema.ConvertTags(ctx, *engineResp.Tags)
		data.Tags, tagDiag = types.ListValueFrom(ctx, data.Tags.ElementType(ctx), tags)
		if tagDiag.HasError() {
			resp.Diagnostics.Append(tagDiag...)
			return
		}
	}

	// Handle spec
	if engineResp.Spec != nil {
		data.Spec, err = helper.JSONObjToNormalized(engineResp.Spec)
		if err != nil {
			resp.Diagnostics.AddError("Response parse Error", fmt.Sprintf("Unable to parse engine spec: %s", err))
			return
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *engineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data engineResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	apiResp, err := r.cl.GetEngineGetWithResponse(ctx, &client.GetEngineGetParams{Id: data.Id.ValueStringPointer()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get engine '%s': %s", data.Id.String(), err))
		return
	}

	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Client Error: %d", apiResp.StatusCode()),
			fmt.Sprintf("Unable to get engine '%s': %s", data.Id.String(), apiResp.Status()),
		)
		return
	}

	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API for engine '%s'", data.Id.String()))
		return
	}

	engineResp := *apiResp.JSON200

	// Update the model with the response data
	helper.ConvertJSONStructToSimpleTF(ctx, engineResp, &data)

	// Handle tags
	var tagDiag diag.Diagnostics
	var tags []internalschema.Tag

	if engineResp.Tags != nil {
		tags = internalschema.ConvertTags(ctx, *engineResp.Tags)
	}

	data.Tags, tagDiag = types.ListValueFrom(ctx, data.Tags.ElementType(ctx), tags)
	if tagDiag.HasError() {
		resp.Diagnostics.Append(tagDiag...)
		return
	}

	// Handle spec
	if engineResp.Spec != nil {
		specJsonStr, err := helper.JSONObjToStr(*engineResp.Spec)
		if err != nil {
			resp.Diagnostics.AddError("Response parse Error", fmt.Sprintf("Unable to parse engine spec for '%s': %s", data.Id.String(), err))
			return
		}

		data.Spec = jsontypes.NewNormalizedValue(specJsonStr)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *engineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data engineResourceModel
	var stateData engineResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	engineConfig := client.EnginesEngineConfig{}

	// Set the basic fields
	err := helper.ConvertTfModelToApiJSON(ctx, data, &engineConfig)
	if err != nil {
		if err != nil {
			resp.Diagnostics.AddError("Error unmarshalling spec", fmt.Sprintf("Update : Unable to unmarshal spec: %s", err))
			return
		}
	}

	// handle the special string kind
	// Convert string to FixedEngineKind
	if !data.Kind.IsNull() {
		kind := data.Kind.ValueString()
		engineConfig.Kind = &kind
	}

	// Handle spec
	if !data.Spec.IsNull() && !data.Spec.IsUnknown() {
		var specObj map[string]interface{}
		err := data.Spec.Unmarshal(&specObj)
		if err != nil {
			resp.Diagnostics.AddError("Error unmarshalling spec", fmt.Sprintf("Update : Unable to unmarshal spec: %s", err))
			return
		}
		engineConfig.Spec = &specObj
	}

	// Handle tags
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []internalschema.Tag
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		apiTags := make([]client.TagsTag, 0, len(tags))
		for _, tag := range tags {
			key := tag.Key.ValueString()
			value := tag.Value.ValueString()
			apiTags = append(apiTags, client.TagsTag{
				Key:   &key,
				Value: &value,
			})
		}
		engineConfig.Tags = &apiTags
	}

	id := stateData.Id.ValueString()
	engineConfig.Id = &id

	// Make the API call
	apiResp, err := r.cl.PostEngineUpdateWithResponse(ctx, engineConfig)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update engine '%s': %s", id, err))
		return
	}

	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Client Error: %d", apiResp.StatusCode()),
			fmt.Sprintf("Unable to update engine '%s': %s", id, apiResp.Status()),
		)
		return
	}

	// After updating the engine, we need to read it to get the full details
	getResp, err := r.cl.GetEngineGetWithResponse(ctx, &client.GetEngineGetParams{Id: &id})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated engine '%s': %s", id, err))
		return
	}

	if getResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Client Error: %d", getResp.StatusCode()),
			fmt.Sprintf("Unable to read updated engine '%s': %s", id, getResp.Status()),
		)
		return
	}

	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API when reading updated engine '%s'", id))
		return
	}

	engineResp := *getResp.JSON200

	// Update the model with the response data
	helper.ConvertJSONStructToSimpleTF(ctx, engineResp, &data)

	// Handle tags
	var tagDiag diag.Diagnostics
	var tags []internalschema.Tag

	if engineResp.Tags != nil {
		tags = internalschema.ConvertTags(ctx, *engineResp.Tags)
	}

	data.Tags, tagDiag = types.ListValueFrom(ctx, data.Tags.ElementType(ctx), tags)
	if tagDiag.HasError() {
		resp.Diagnostics.Append(tagDiag...)
		return
	}

	// Handle spec
	if engineResp.Spec != nil {
		specJsonStr, err := helper.JSONObjToStr(*engineResp.Spec)
		if err != nil {
			resp.Diagnostics.AddError("Response parse Error", fmt.Sprintf("Unable to parse engine spec for '%s': %s", id, err))
			return
		}

		data.Spec = jsontypes.NewNormalizedValue(specJsonStr)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *engineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data engineResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	apiResp, err := r.cl.DeleteEngineDeleteWithResponse(ctx, &client.DeleteEngineDeleteParams{Id: data.Id.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete engine '%s': %s", data.Id.String(), err))
		return
	}

	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Client Error: %d", apiResp.StatusCode()),
			fmt.Sprintf("Unable to delete engine '%s': %s", data.Id.String(), apiResp.Status()),
		)
		return
	}
}
