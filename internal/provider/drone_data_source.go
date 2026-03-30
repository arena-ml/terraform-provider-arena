// Copyright (c) ArenaML Labs Pvt Ltd.

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/arena-ml/terraform-provider-arenaml/generator/client"
	"github.com/arena-ml/terraform-provider-arenaml/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kr/pretty"
)

type droneDataSource struct {
	cl *client.ClientWithResponses
}

var _ datasource.DataSource = (*droneDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*droneDataSource)(nil)

func NewDroneDataSource() datasource.DataSource {
	return &droneDataSource{}
}

func (d *droneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_drone"
}

func (d *droneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.DroneDSchema()
}

func (d *droneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cl, ok := req.ProviderData.(*client.ClientWithResponses)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *oapi client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.cl = cl
}

func (d *droneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data schema.DroneModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("unable to read drone tf spec %+v", resp.Diagnostics.Errors()))
		return
	}

	if data.ID.IsNull() || data.ID.IsUnknown() {
		resp.Diagnostics.AddError("id cannot be null for this datasource", "id cannot be null for this datasource")
		return
	}

	apiResp, err := d.cl.GetDronesGetWithResponse(ctx, &client.GetDronesGetParams{Id: data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API client error in GET Drone: id: %s \nerr: %s", data.ID.String(), err))
		return
	}

	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(fmt.Sprintf("Client Error : %d", apiResp.StatusCode()), fmt.Sprintf("Unable to get drone '%s\n code : %d'",
			data.ID.String(), apiResp.StatusCode()))
		return
	}

	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error : null data in response", fmt.Sprintf("API Call error \nid : '%s'", data.ID.String()))
		return
	}

	droneNode := apiResp.JSON200

	tflog.Error(ctx, fmt.Sprintf("API Call Error : \n\n%s\n\n", pretty.Sprint(droneNode)))

	err = data.FillFromResp(ctx, *droneNode)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API response parsing error \nid : '%s' , err: %s", data.ID.String(), err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
