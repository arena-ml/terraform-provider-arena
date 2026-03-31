// Copyright (c) ArenaML Labs Pvt Ltd.

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/arena-ml/terraform-provider-arenaml/generator/client"
	"github.com/arena-ml/terraform-provider-arenaml/helper"
	"github.com/arena-ml/terraform-provider-arenaml/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kr/pretty"
)

var _ datasource.DataSource = (*engineDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*engineDataSource)(nil)

const suffixClusterManager = "cluster_manager"

func NewEngineDataSource() datasource.DataSource {
	return &engineDataSource{}
}

type engineDataSource struct {
	cl *client.ClientWithResponses
}

type engineDataSourceModel struct {
	ID types.String `tfsdk:"id"`
}

func (d *engineDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + suffixClusterManager
}

func (d *engineDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.EngineDataSourceSchema(ctx)
}

func (d *engineDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cl, ok := req.ProviderData.(*client.ClientWithResponses)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *oapi cleint, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.cl = cl
}

func (d *engineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data schema.EngineModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	apiResp, err := d.cl.GetEngineGetWithResponse(ctx, &client.GetEngineGetParams{Id: data.Id.ValueStringPointer()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error : "+err.Error(), fmt.Sprintf("Unable to get engine '%s'", data.Id.String()))
		return
	}

	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error : %d", apiResp.StatusCode()), fmt.Sprintf("Unable to get engine '%s'", data.Id.String()))
		return
	}

	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error : null data in response", fmt.Sprintf("id : '%s'", data.Id.String()))
		return
	}

	engineResp := *apiResp.JSON200

	tflog.Warn(ctx, pretty.Sprint(engineResp))

	// copy the basic values from resp to data
	helper.ConvertJSONStructToSimpleTF(ctx, *apiResp.JSON200, &data)
	var tagDiag diag.Diagnostics

	var tags []schema.Tag

	if engineResp.Tags != nil {
		tags = schema.ConvertTags(ctx, *engineResp.Tags)
	}

	data.Tags, tagDiag = types.ListValueFrom(ctx, data.Tags.ElementType(ctx), tags)
	if tagDiag.HasError() {
		resp.Diagnostics.Append(tagDiag...)
		return
	}

	if engineResp.Spec != nil {
		specJsonStr, err := helper.JSONObjToStr(*engineResp.Spec)
		if err != nil {
			resp.Diagnostics.AddError("Response parse Error : "+err.Error(), fmt.Sprintf("Unable to get engine '%s'", data.Id.String()))
			return
		}

		data.Spec = jsontypes.NewNormalizedValue(specJsonStr)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
