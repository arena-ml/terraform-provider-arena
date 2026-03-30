// Copyright (c) ArenaML Labs Pvt Ltd.

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/arena-ml/terraform-provider-arenaml/generator/client"
	"github.com/arena-ml/terraform-provider-arenaml/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kr/pretty"
)

type inputDataSource struct {
	cl *client.ClientWithResponses
}

var _ datasource.DataSource = (*inputDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*inputDataSource)(nil)

type inputDataSourceConfig struct {
	ID types.String `tfsdk:"id"`
}

func NewInputDatasource() datasource.DataSource {
	return &inputDataSource{}
}

func (i *inputDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_input"
}

func (i *inputDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.InputNodeDSchema()
}

func (i *inputDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	cl, ok := request.ProviderData.(*client.ClientWithResponses)

	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *oapi cleint, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}
	i.cl = cl
}

func (i *inputDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data schema.InputNode

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("unable to read input tf spec %+v", resp.Diagnostics.Errors()))
		return
	}

	if data.ID.IsNull() || data.ID.IsUnknown() {
		resp.Diagnostics.AddError("id cannot be null for this datasource", "id cannot be null for this datasource")
	}

	apiResp, err := i.cl.GetPipelineNodesOneWithResponse(ctx, &client.GetPipelineNodesOneParams{Id: data.ID.ValueString(), Kind: "input"})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API client error in GET Input: id: %s \nerr: %s", data.ID.String(), err))
		return
	}

	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error : %d", apiResp.StatusCode()), fmt.Sprintf("Unable to get input '%s\n code : %d'",
			data.ID.String(), apiResp.StatusCode()))
		return
	}

	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error : null data in response", fmt.Sprintf("API Call error \nid : '%s'", data.ID.String()))
		return
	}

	inputs := apiResp.JSON200.Inputs
	if inputs == nil || len(*inputs) != 1 {
		resp.Diagnostics.AddError("unexpected response body", "response should have exactly one input node")
		return
	}

	inputNode := (*inputs)[0]

	tflog.Error(ctx, fmt.Sprintf("API Call Error : \n\n%s\n\n", pretty.Sprint(inputNode)))
	// copy the basic values from resp to data
	err = data.FillFromResp(ctx, inputNode)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error \nid : '%s' , err: %s", data.ID.String(), err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
