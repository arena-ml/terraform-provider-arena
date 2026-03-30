// Copyright (c) ArenaML Labs Pvt Ltd.

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/arena-ml/terraform-provider-arenaml/generator/client"
	"github.com/arena-ml/terraform-provider-arenaml/helper"
	internalschema "github.com/arena-ml/terraform-provider-arenaml/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure implementation satisfies interfaces
var _ resource.Resource = (*basisResource)(nil)
var _ resource.ResourceWithConfigure = (*basisResource)(nil)

func NewBasisResource() resource.Resource { return &basisResource{} }

type basisResource struct {
	cl *client.ClientWithResponses
}

func (r *basisResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
// (holds top-level fields and nested config sections)
type basisResourceModel = internalschema.BasisModel

func (r *basisResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_basis"
}

func (r *basisResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Resource schema: mirrors Basis model. Some attributes are computed from server.
	resp.Schema = internalschema.BasisResourceSchema()
}

func (r *basisResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data basisResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entBasis, err := r.tfToAPIData(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("unable to create Basis", "unable to create Basis : "+err.Error())
		return
	}
	// Call Create
	apiResp, err := r.cl.PostBasisCreateWithResponse(ctx, *entBasis)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create basis: %s", err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to create basis: %s", apiResp.Status()))
		return
	}

	// Parse response to get ID
	created := apiResp.JSON200

	if created.Id == nil {
		resp.Diagnostics.AddError("Client Error", "Created basis has no ID")
		return
	}

	// Read back full entity
	getResp, err := r.cl.GetBasisGetWithResponse(ctx, &client.GetBasisGetParams{Id: created.Id})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created basis: %s", err))
		return
	}
	if getResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", getResp.StatusCode()), fmt.Sprintf("Unable to read created basis: %s", getResp.Status()))
		return
	}
	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Received nil response from API when reading created basis")
		return
	}

	// Fill TF model from response
	if err := data.FillFromResp(ctx, *getResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *basisResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data basisResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.cl.GetBasisGetWithResponse(ctx, &client.GetBasisGetParams{Id: data.ID.ValueStringPointer()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get basis '%s': %s", data.ID.String(), err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to get basis '%s': %s", data.ID.String(), apiResp.Status()))
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API for basis '%s'", data.ID.String()))
		return
	}

	if err := data.FillFromResp(ctx, *apiResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *basisResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data basisResourceModel
	var stateData basisResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiBasis, err := r.tfToAPIData(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("unable to create Basis", "unable to create Basis : "+err.Error())
		return
	}
	id := stateData.ID.ValueString()
	apiBasis.Id = &id

	apiResp, err := r.cl.PostBasisUpdateWithResponse(ctx, *apiBasis)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update basis '%s': %s", id, err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to update basis '%s': %s", id, apiResp.Status()))
		return
	}

	getResp, err := r.cl.GetBasisGetWithResponse(ctx, &client.GetBasisGetParams{Id: &id})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated basis '%s': %s", id, err))
		return
	}
	if getResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", getResp.StatusCode()), fmt.Sprintf("Unable to read updated basis '%s': %s", id, getResp.Status()))
		return
	}
	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Received nil response from API when reading updated basis '%s'", id))
		return
	}
	if err := data.FillFromResp(ctx, *getResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *basisResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data basisResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.cl.DeleteBasisDeleteWithResponse(ctx, &client.DeleteBasisDeleteParams{Id: data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete basis '%s': %s", data.ID.String(), err))
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error: %d", apiResp.StatusCode()), fmt.Sprintf("Unable to delete basis '%s': %s", data.ID.String(), apiResp.Status()))
		return
	}
}
func (r *basisResource) tfToAPIData(ctx context.Context, data basisResourceModel) (entBasis *client.EntBasis, err error) {
	// parse api response
	entBasis = &client.EntBasis{}

	err = helper.ConvertTfModelToApiJSON(ctx, data, entBasis)
	if err != nil {
		return
	}
	// convert ModelCommon fields
	err = helper.ConvertTfModelToApiJSON(ctx, data.ModelCommon, entBasis)
	if err != nil {
		return
	}

	cfg := client.SchemaBasisConfig{}

	// fill primitive tf fields of the basis source field
	basisSrc := &client.SchemaBasisSource{}
	basisWatcher := &client.SchemaBasisWatcher{}
	basisCollector := &client.SchemaBasisCollector{}

	err = helper.ConvertTfModelToApiJSON(ctx, data.Source, basisSrc)
	if err != nil {
		return
	}
	cfg.Source = basisSrc

	// allowed_envs
	if !data.AllowedEnvs.IsNull() && !data.AllowedEnvs.IsUnknown() {
		allowedEnv, ok := helper.TfListStrToGoSlice(ctx, data.AllowedEnvs)
		if !ok {
			return nil, fmt.Errorf("failed to parse allow_envs %s", data.AllowedEnvs)
		}
		cfg.AllowedEnvs = &allowedEnv
	}

	// watcher
	if data.Watcher != nil {
		err = helper.ConvertTfModelToApiJSON(ctx, data.Watcher, basisWatcher)
		if err != nil {
			return
		}

		// parse env map
		if !data.Watcher.Env.IsNull() && !data.Watcher.Env.IsUnknown() {
			env, ok := helper.TfMapStrToGoMap(ctx, data.Watcher.Env)
			if !ok {
				return nil, fmt.Errorf("failed to parse env %s", data.Watcher.Env)
			}
			basisWatcher.Env = &env
		}

		// parse registry_auth
		if !data.Watcher.RegistryAuth.IsNull() && !data.Watcher.RegistryAuth.IsUnknown() {
			ra, ok := helper.TfMapStrToGoMap(ctx, data.Watcher.RegistryAuth)
			if !ok {
				return
			}
			basisWatcher.RegistryAuth = &ra
		}
		// parse run_spec
		if !data.Watcher.RunSpec.IsNull() && !data.Watcher.RunSpec.IsUnknown() {
			spec, err := helper.TfJSONToGoMapInterface(ctx, data.Watcher.RunSpec)
			if err != nil {
				return nil, fmt.Errorf("failed to parse run spec of watcher :\n %s %s", data.Watcher.RunSpec, err.Error())
			}
			basisWatcher.RunSpec = &spec
		}

		cfg.Watcher = basisWatcher
	}

	// collector
	if data.Collector != nil {
		err = helper.ConvertTfModelToApiJSON(ctx, data.Collector, basisCollector)
		if err != nil {
			return
		}

		if !data.Collector.RunSpec.IsNull() && !data.Collector.RunSpec.IsUnknown() {
			var runSpec map[string]interface{}
			runSpec, err = helper.TfJSONToGoMapInterface(ctx, data.Collector.RunSpec)
			if err != nil {
				return nil, fmt.Errorf("failed to parse run spec of collector :\n %s %s", data.Watcher.RunSpec, err.Error())
			}
			basisCollector.RunSpec = &runSpec
		}

		cfg.Collector = basisCollector
	}

	entBasis.Config = &cfg

	return entBasis, nil
}
