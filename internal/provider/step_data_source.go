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

type stepDataSource struct {
	cl *client.ClientWithResponses
}

var _ datasource.DataSource = (*stepDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*stepDataSource)(nil)

type stepDataSourceConfig struct {
	ID types.String `tfsdk:"id"`
}

func NewStepDatasource() datasource.DataSource {
	return &stepDataSource{}
}

func (s *stepDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_step"
}

func (s *stepDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.NodeStepDSchema()
}

func (s *stepDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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
	s.cl = cl
}

func (s *stepDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data schema.NodeStep

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("unable to read step tf spec %+v", resp.Diagnostics.Errors()))
		return
	}

	if data.ID.IsNull() || data.ID.IsUnknown() {
		resp.Diagnostics.AddError("id cannot be null for this datasource", "id cannot be null for this datasource")
	}

	apiResp, err := s.cl.GetPipelineNodesOneWithResponse(ctx, &client.GetPipelineNodesOneParams{Id: data.ID.ValueString(), Kind: "step"})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API client error in GET Step: id: %s \nerr: %s", data.ID.String(), err))
		return
	}

	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error : %d", apiResp.StatusCode()), fmt.Sprintf("Unable to get step '%s\n code : %d'",
			data.ID.String(), apiResp.StatusCode()))
		return
	}

	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error : null data in response", fmt.Sprintf("API Call error \nid : '%s'", data.ID.String()))
		return
	}

	steps := apiResp.JSON200.Steps
	if steps == nil || len(*steps) != 1 {
		resp.Diagnostics.AddError("unexpected response body", "response should have exactly one step node")
		return
	}

	stepNode := (*steps)[0]

	tflog.Error(ctx, fmt.Sprintf("API Call Error : \n\n%s\n\n", pretty.Sprint(stepNode)))
	// copy the basic values from resp to data
	err = data.FillFromResp(ctx, stepNode)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error \nid : '%s' , err: %s", data.ID.String(), err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
