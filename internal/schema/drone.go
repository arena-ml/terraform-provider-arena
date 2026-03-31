// Copyright (c) ArenaML Labs Pvt Ltd.

package schema

import (
	"context"
	"fmt"

	"github.com/arena-ml/terraform-provider-arenaml/generator/client"
	"github.com/arena-ml/terraform-provider-arenaml/helper"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type DroneModel struct {
	ModelCommon
	DroneCommon
	ProfileID types.String `tfsdk:"profile_id"`
	SwarmID   types.String `tfsdk:"swarm_id"`
	Inactive  types.Bool   `tfsdk:"inactive"`
}

type DroneProfileModel struct {
	ModelCommon
	DroneCommon
}

type DroneCommon struct {
	Kind  types.String         `tfsdk:"kind"`
	Spec  *DroneSpec           `tfsdk:"spec"`
	Links jsontypes.Normalized `tfsdk:"links"`
}

type DroneSpec struct {
	Arch       types.String `tfsdk:"arch"`
	Compute    types.Map    `tfsdk:"compute"`
	Details    types.Map    `tfsdk:"details"`
	Gpu        types.Map    `tfsdk:"gpu"`
	Storage    types.Map    `tfsdk:"storage"`
	MemoryInGB types.Int32  `tfsdk:"memory_in_gb"`
	Model      types.Map    `tfsdk:"model"`
	Networks   types.Map    `tfsdk:"networks"`
	Npu        types.Map    `tfsdk:"npu"`
	Power      types.Map    `tfsdk:"power"`
}

type DroneStorage struct {
	GUID      types.String `tfsdk:"guid"`
	DevPath   types.String `tfsdk:"dev_path"`
	MountPath types.String `tfsdk:"mount_path"`
	Capacity  types.Int32  `tfsdk:"capacity"`
	Kind      types.String `tfsdk:"kind"`
}

func droneStorageAttrs() []BaseSchema {
	// return base schema attrs for the model DroneDeviceStorage
	return []BaseSchema{
		{Name: "guid", AttrType: TfString, Optional: true, Desc: "disk guid"},
		{Name: "dev_path", AttrType: TfString, Optional: true, Desc: "path under /dev tree"},
		{Name: "mount_path", AttrType: TfString, Optional: true, Desc: "filesystem mount path of this disk"},
		{Name: "capacity", AttrType: TfInt, Optional: true, Desc: "storage capacity in GB"},
		{Name: "kind", AttrType: TfString, Optional: true, Desc: "kind of storage e.g. ssd, hdd, nvme, flash, sd-card"},
	}
}

func (ds DroneStorage) objType() types.ObjectType {
	attrMap := make(map[string]attr.Type)
	bs := droneStorageAttrs()
	for _, s := range bs {
		attrMap[s.Name] = s.TFAttrType()
	}
	return types.ObjectType{
		AttrTypes: attrMap,
	}
}

type DroneDeviceNetwork struct {
	Kind       types.String  `tfsdk:"kind"`
	MAC        types.String  `tfsdk:"mac"`
	Bandwidth  types.Float64 `tfsdk:"bandwidth"`
	MaxRange   types.Float64 `tfsdk:"max_range"`
	PowerUsage types.Float64 `tfsdk:"power_usage"`
}

func (dn DroneDeviceNetwork) objType() types.ObjectType {
	attrMap := make(map[string]attr.Type)
	bs := droneNetworkAttrs()
	for _, s := range bs {
		attrMap[s.Name] = s.TFAttrType()
	}
	return types.ObjectType{
		AttrTypes: attrMap,
	}
}

func droneNetworkAttrs() []BaseSchema {
	return []BaseSchema{
		{Name: "bandwidth", AttrType: TFFloat64, Optional: true, Desc: "network bandwidth in mbps at mid distance from max"},
		{Name: "kind", AttrType: TfString, Optional: true, Desc: "kind of network interface radio/wifi5 etc"},
		{Name: "mac", AttrType: TfString, Optional: true, Desc: "MAC address"},
		{Name: "max_range", AttrType: TFFloat64, Optional: true, Desc: "maximum range in meters"},
		{Name: "power_usage", AttrType: TFFloat64, Optional: true, Desc: "power usage in milli watts"},
	}
}

func (d *DroneModel) FillFromResp(ctx context.Context, r client.EntDrone) (err error) {
	helper.ConvertJSONStructToSimpleTF(ctx, r, d)

	mc := &ModelCommon{}
	helper.ConvertJSONStructToSimpleTF(ctx, r, mc)
	d.ModelCommon = *mc

	dc := &DroneCommon{}
	err = dc.FillFromResp(ctx, r.Kind, r.Spec, r.Links)
	if err != nil {
		return
	}
	d.DroneCommon = *dc

	if r.Edges != nil {
		if r.Edges.Profile != nil && r.Edges.Profile.Id != nil {
			d.ProfileID = types.StringValue(*r.Edges.Profile.Id)
		}
		if r.Edges.Swarm != nil && r.Edges.Swarm.Id != nil {
			d.SwarmID = types.StringValue(*r.Edges.Swarm.Id)
		}
	}

	return
}

func (d *DroneProfileModel) FillFromResp(ctx context.Context, r client.EntDroneProfile) (err error) {
	mc := &ModelCommon{}
	helper.ConvertJSONStructToSimpleTF(ctx, r, mc)
	d.ModelCommon = *mc

	dc := &DroneCommon{}
	err = dc.FillFromResp(ctx, r.Kind, r.Spec, r.Links)
	if err != nil {
		return
	}
	d.DroneCommon = *dc

	return
}

func (dc *DroneCommon) FillFromResp(ctx context.Context, kind *string, rSpec *client.SchemaDroneSpec, links *client.SchemaDroneLinks) (err error) {
	if kind != nil {
		dc.Kind = types.StringValue(*kind)
	}

	if rSpec != nil {
		spec := &DroneSpec{}
		err = spec.FillFromResp(ctx, *rSpec)
		if err != nil {
			return
		}
		dc.Spec = spec
	}

	if links != nil {
		dc.Links, err = helper.JSONObjToNormalized(links)
		if err != nil {
			return fmt.Errorf("links: %w", err)
		}
	} else {
		dc.Links = jsontypes.NewNormalizedNull()
	}

	return nil
}

func (s *DroneSpec) FillFromResp(ctx context.Context, resp client.SchemaDroneSpec) (err error) {
	helper.ConvertJSONStructToSimpleTF(ctx, resp, s)

	s.Compute, err = helper.TfMapFromGoMapFloat32(ctx, resp.Compute)
	if err != nil {
		return fmt.Errorf("compute: %w", err)
	}

	s.Details, err = helper.ToTfStringMap(ctx, resp.Details)
	if err != nil {
		return fmt.Errorf("details: %w", err)
	}

	s.Gpu, err = helper.TfMapFromGoMapFloat32(ctx, resp.Gpu)
	if err != nil {
		return fmt.Errorf("gpu: %w", err)
	}

	s.Model, err = helper.ToTfStringMap(ctx, resp.Model)
	if err != nil {
		return fmt.Errorf("model: %w", err)
	}

	s.Npu, err = helper.TfMapFromGoMapFloat32(ctx, resp.Npu)
	if err != nil {
		return fmt.Errorf("npu: %w", err)
	}

	s.Power, err = helper.ToTfStringMap(ctx, resp.Power)
	if err != nil {
		return fmt.Errorf("power: %w", err)
	}

	if resp.Storage != nil {
		storageTF := make(map[string]DroneStorage)
		helper.MapConvertJsonToTf(ctx, *resp.Storage, storageTF)
		tflog.Error(ctx, fmt.Sprintf("\n\n\n%+v\n\n", storageTF))

		var dg diag.Diagnostics
		s.Storage, dg = types.MapValueFrom(ctx, DroneStorage{}.objType(), &storageTF)
		if dg != nil && dg.HasError() {
			return fmt.Errorf("local_storage: %w", err)
		}
	} else {
		s.Storage = types.MapNull(DroneStorage{}.objType())
	}

	if resp.Networks != nil {
		networkTf := make(map[string]DroneDeviceNetwork)
		helper.MapConvertJsonToTf(ctx, *resp.Networks, networkTf)

		var dg diag.Diagnostics
		s.Networks, dg = types.MapValueFrom(ctx, DroneDeviceNetwork{}.objType(), &networkTf)
		if dg != nil && dg.HasError() {
			return fmt.Errorf("unable to parse network spec")
		}
	} else {
		s.Networks = types.MapNull(DroneDeviceNetwork{}.objType())
	}

	// if resp.Networks != nil {
	// 	// s.Networks, err = helper.JSONObjToNormalized(resp.Networks)
	// 	s.Networks, _ = types.MapValueFrom(ctx, types.ObjectType{}, resp.Networks)
	// 	// if err != nil {
	// 	// 	return fmt.Errorf("networks: %w", err)
	// 	// }
	// }

	return nil
}

func (d *DroneModel) ToModelJSON(ctx context.Context) (jsonDrone client.EntDrone, err error) {
	err = helper.ConvertTfModelToApiJSON(ctx, d, &jsonDrone)
	if err != nil {
		return
	}

	if d.Spec != nil {
		spec, sErr := d.Spec.ToModelJSON(ctx)
		if sErr != nil {
			err = sErr
			return
		}
		jsonDrone.Spec = &spec
	}

	if !d.Links.IsNull() && !d.Links.IsUnknown() {
		var links client.SchemaDroneLinks
		diag := d.Links.Unmarshal(&links)
		if diag.HasError() {
			err = fmt.Errorf("%s", diag.Errors()[0].Detail())
			return
		}
		jsonDrone.Links = &links
	}
	jsonDrone.Edges = &client.EntDroneEdges{}

	if !d.ProfileID.IsNull() && !d.ProfileID.IsUnknown() {
		id := d.ProfileID.ValueString()
		jsonDrone.Edges.Profile = &client.EntDroneProfile{Id: &id}
	}

	if !d.SwarmID.IsNull() && !d.SwarmID.IsUnknown() {
		id := d.SwarmID.ValueString()
		jsonDrone.Edges.Swarm = &client.EntSwarm{Id: &id}
	}

	return
}

func (d *DroneProfileModel) ToModelJSON(ctx context.Context) (jsonProfile client.EntDroneProfile, err error) {
	err = helper.ConvertTfModelToApiJSON(ctx, d.ModelCommon, &jsonProfile)
	if err != nil {
		return
	}

	err = helper.ConvertTfModelToApiJSON(ctx, d.DroneCommon, &jsonProfile)
	if err != nil {
		return
	}

	if d.Spec != nil {
		spec, sErr := d.Spec.ToModelJSON(ctx)
		if sErr != nil {
			err = sErr
			return
		}
		jsonProfile.Spec = &spec
	}

	if !d.Links.IsNull() && !d.Links.IsUnknown() {
		jsonProfile.Links, err = d.giveJsonLinks()
		if err != nil {
			return
		}
	}

	return
}

func (s *DroneSpec) ToModelJSON(ctx context.Context) (spec client.SchemaDroneSpec, err error) {
	err = helper.ConvertTfModelToApiJSON(ctx, *s, &spec)

	if !s.Compute.IsNull() && !s.Compute.IsUnknown() {
		compute, ok := helper.GoMapFloat32FromTfMap(ctx, s.Compute)
		if !ok {
			err = fmt.Errorf("failed to convert compute map")
			return
		}
		spec.Compute = &compute
	}

	if !s.Details.IsNull() && !s.Details.IsUnknown() {
		details, ok := helper.TfMapStrToGoMap(ctx, s.Details)
		if !ok {
			err = fmt.Errorf("failed to convert details map")
			return
		}
		spec.Details = &details
	}

	if !s.Gpu.IsNull() && !s.Gpu.IsUnknown() {
		gpu, ok := helper.GoMapFloat32FromTfMap(ctx, s.Gpu)
		if !ok {
			err = fmt.Errorf("failed to convert gpu map")
			return
		}
		spec.Gpu = &gpu
	}

	if !s.Model.IsNull() && !s.Model.IsUnknown() {
		model, ok := helper.TfMapStrToGoMap(ctx, s.Model)
		if !ok {
			err = fmt.Errorf("failed to convert model map")
			return
		}
		spec.Model = &model
	}

	if !s.Npu.IsNull() && !s.Npu.IsUnknown() {
		npu, ok := helper.GoMapFloat32FromTfMap(ctx, s.Npu)
		if !ok {
			err = fmt.Errorf("failed to convert npu map")
			return
		}
		spec.Npu = &npu
	}

	if !s.Power.IsNull() && !s.Power.IsUnknown() {
		power, ok := helper.TfMapStrToGoMap(ctx, s.Power)
		if !ok {
			err = fmt.Errorf("failed to convert power map")
			return
		}
		spec.Power = &power
	}

	if !s.Storage.IsNull() && !s.Storage.IsUnknown() {
		storageTF := make(map[string]DroneStorage)
		// diag := s.LocalStorage.Unmarshal(&lsJSON)
		diag := s.Storage.ElementsAs(ctx, &storageTF, false)
		if diag.HasError() {
			err = fmt.Errorf("%s", diag.Errors()[0].Detail())
			return
		}
		// spec.LocalStorage = &lsJSON
		storageJSON := make(map[string]client.SchemaDeviceStorage)
		for k, vTF := range storageTF {
			vJSON := client.SchemaDeviceStorage{}
			err = helper.ConvertTfModelToApiJSON(ctx, vTF, &vJSON)
			if err != nil {
				return
			}
			storageJSON[k] = vJSON
		}

		spec.Storage = &storageJSON
	}

	if !s.Networks.IsNull() && !s.Networks.IsUnknown() {
		// netTF := make(map[string]DroneDeviceNetwork)
		// // diag := s.Networks.Unmarshal(&netJSON)
		// diag := s.Networks.ElementsAs(ctx, &netTF, false)
		// if diag.HasError() {
		// 	err = fmt.Errorf(diag.Errors()[0].Detail())
		// 	return
		// }
		// netJSON := make(map[string]client.SchemaDeviceNetwork)
		// for k, vTF := range netTF {
		// 	vJSON := client.SchemaDeviceNetwork{}
		// 	err = helper.ConvertTfModelToApiJSON(ctx, vTF, &vJSON)
		// 	if err != nil {
		// 		return
		// 	}
		// 	netJSON[k] = vJSON
		// }

		var netJSON map[string]client.SchemaDeviceNetwork

		netJSON, err = convertTtoG(ctx, s.Networks, DroneDeviceNetwork{}, client.SchemaDeviceNetwork{})
		if err != nil {
			return
		}

		spec.Networks = &netJSON
	}

	return
}

func convertTtoG[S any, T any](ctx context.Context, tm types.Map, src S, trg T) (netJSON map[string]T, err error) {
	if !tm.IsNull() && !tm.IsUnknown() {
		netTF := make(map[string]S)
		// diag := s.Networks.Unmarshal(&netJSON)
		diag := tm.ElementsAs(ctx, &netTF, false)
		if diag.HasError() {
			err = fmt.Errorf("%s", diag.Errors()[0].Detail())
			return
		}
		netJSON = make(map[string]T)
		for k, vTF := range netTF {
			var vJSON T
			err = helper.ConvertTfModelToApiJSON(ctx, vTF, &vJSON)
			if err != nil {
				return
			}
			netJSON[k] = vJSON
		}

	}
	return
}

func (dc *DroneCommon) giveJsonLinks() (links *client.SchemaDroneLinks, err error) {
	// return nil values of value not present
	if dc.Links.IsNull() || dc.Links.IsUnknown() {
		return
	}

	diag := dc.Links.Unmarshal(links)
	if diag.HasError() {
		err = fmt.Errorf("%s", diag.Errors()[0].Detail())
	}
	return
}

func droneSpecAttrs() []BaseSchema {
	return []BaseSchema{
		{Name: "arch", AttrType: TfString, Required: true, Desc: "CPU architecture (e.g. amd64, arm64)"},
		{Name: "compute", AttrType: TfMap, SubType: TFFloat, Optional: true, Desc: "map of core to frequency"},
		{Name: "details", AttrType: TfMap, SubType: TfString, Optional: true, Desc: "any other details"},
		{Name: "gpu", AttrType: TfMap, SubType: TFFloat, Optional: true, Desc: "GPU parameters"},
		{Name: "memory_in_gb", AttrType: TfInt, Required: true, Desc: "memory in GB"},
		{Name: "model", AttrType: TfMap, SubType: TfString, Optional: true, Desc: "model parameters"},
		{Name: "storage", AttrType: TfMapNested, Optional: true, Desc: "map of local storage options", NestedAttrs: droneStorageAttrs()},
		{Name: "networks", AttrType: TfMapNested, Optional: true, Desc: "network interfaces", NestedAttrs: droneNetworkAttrs()},
		{Name: "npu", AttrType: TfMap, SubType: TFFloat, Optional: true, Desc: "NPU parameters"},
		{Name: "power", AttrType: TfMap, SubType: TfString, Optional: true, Desc: "power related details"},
	}
}

func droneAttrs() []BaseSchema {
	attrs := giveCommonAttributes()
	droneSpecific := []BaseSchema{
		{Name: "kind", AttrType: TfString, Optional: true, Desc: "drone kind"},
		{Name: "links", AttrType: TfJSON, Optional: true, Desc: "device interface links"},
		{Name: "profile_id", AttrType: TfString, Optional: true, Desc: "id of drone profile"},
		{Name: "swarm_id", AttrType: TfString, Optional: true, Desc: "id of swarm the drone belongs to"},
		{Name: "inactive", AttrType: TfBoolean, Optional: true, Desc: "whether the drone is inactive"},
	}
	return append(attrs, droneSpecific...)
}

func droneProfileAttrs() []BaseSchema {
	attrs := giveCommonAttributes()
	droneProfileSpecific := []BaseSchema{
		{Name: "kind", AttrType: TfString, Optional: true, Desc: "drone kind"},
		{Name: "links", AttrType: TfJSON, Optional: true, Desc: "device interface links"},
	}
	return append(attrs, droneProfileSpecific...)
}

func DroneResourceSchema() rschema.Schema {
	attrs := ResAttributes(droneAttrs())
	attrs["spec"] = rschema.SingleNestedAttribute{
		Attributes:  ResAttributes(droneSpecAttrs()),
		Required:    true,
		Description: "drone device specification",
	}
	return rschema.Schema{
		Attributes:          attrs,
		Description:         "Manages a drone resource.",
		MarkdownDescription: "Manages a drone resource.",
	}
}

func DroneProfileResourceSchema() rschema.Schema {
	attrs := ResAttributes(droneProfileAttrs())
	attrs["spec"] = rschema.SingleNestedAttribute{
		Attributes:  ResAttributes(droneSpecAttrs()),
		Required:    true,
		Description: "drone device specification",
	}
	return rschema.Schema{
		Attributes:          attrs,
		Description:         "Manages a drone profile resource.",
		MarkdownDescription: "Manages a drone profile resource.",
	}
}

func DroneDSchema() dschema.Schema {
	attrs := DSAttributes(droneAttrs())
	attrs["spec"] = dschema.SingleNestedAttribute{
		Attributes:  DSAttributes(droneSpecAttrs()),
		Computed:    true,
		Description: "drone device specification",
	}
	return dschema.Schema{
		Attributes:          attrs,
		Description:         "Fetches a drone by ID.",
		MarkdownDescription: "Fetches a drone by ID.",
	}
}

func DroneProfileDSchema() dschema.Schema {
	attrs := DSAttributes(droneProfileAttrs())
	attrs["spec"] = dschema.SingleNestedAttribute{
		Attributes:  DSAttributes(droneSpecAttrs()),
		Computed:    true,
		Description: "drone device specification",
	}
	return dschema.Schema{
		Attributes:          attrs,
		Description:         "Fetches a drone profile by ID.",
		MarkdownDescription: "Fetches a drone profile by ID.",
	}
}
