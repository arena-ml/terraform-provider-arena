package schema

import (
	"context"
	"fmt"

	"github.com/arena-ml/terraform-provider-arenaml/generator/client"
	"github.com/arena-ml/terraform-provider-arenaml/helper"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type SensorProfileModel struct {
	SensorCommon
	ModelCommon
}

type SensorModel struct {
	ProfileID types.String `tfsdk:"profile_id"` // id of sensor profile
	Inactive  types.Bool   `tfsdk:"inactive"`
	SensorCommon
	ModelCommon
}

type SensorCommon struct {
	Kind      types.String     `tfsdk:"kind"`
	Spec      *SensorSpec      `tfsdk:"spec"`
	Interface *SensorInterface `tfsdk:"interface"`
}

type SensorSpec struct {
	Comm        types.Map            `tfsdk:"comm"`
	HFov        types.Float32        `tfsdk:"h_fov"`
	MaxRange    types.Float32        `tfsdk:"max_range"`
	MaxRateInHz types.Float32        `tfsdk:"max_rate_in_hz"`
	Media       types.Map            `tfsdk:"media"`
	MinRange    types.Float32        `tfsdk:"min_range"`
	MinRateInHz types.Float32        `tfsdk:"min_rate_in_hz"`
	Misc        jsontypes.Normalized `tfsdk:"misc"`
	Model       types.Map            `tfsdk:"model"`
	Operating   types.Map            `tfsdk:"operating"`
	Power       types.Map            `tfsdk:"power"`
	RangeUnit   types.String         `tfsdk:"range_unit"`
	Units       types.String         `tfsdk:"units"`
	VFoV        types.Float32        `tfsdk:"v_fov"`
}

type SensorInterface struct {
	Cable    types.Map    `tfsdk:"cable"`
	DeviceIo types.Map    `tfsdk:"device_io"`
	Kind     types.String `tfsdk:"kind"`
	SensorIo types.Map    `tfsdk:"sensor_io"`
}

func (s *SensorProfileModel) FillFromResp(ctx context.Context, r client.EntSensorProfile) (err error) {
	helper.ConvertJSONStructToSimpleTF(ctx, r, s)
	mc := &ModelCommon{}
	helper.ConvertJSONStructToSimpleTF(ctx, r, mc)
	s.ModelCommon = *mc
	tflog.Warn(ctx, fmt.Sprintf("model common for sensor profile \n\n %+v \n\n %+v \n\n", mc, r))

	sc := &SensorCommon{}
	err = sc.FillFromResp(ctx, r.Kind, r.Spec, r.Interface)
	if err != nil {
		return
	}
	s.SensorCommon = *sc

	return
}

func (s *SensorProfileModel) ToModelJSON(ctx context.Context) (jsonProfile client.EntSensorProfile, err error) {
	err = helper.ConvertTfModelToApiJSON(ctx, s.ModelCommon, &jsonProfile)
	if err != nil {
		return
	}

	err = helper.ConvertTfModelToApiJSON(ctx, s.SensorCommon, &jsonProfile)
	if err != nil {
		return
	}

	if s.Spec != nil {
		spec, sErr := s.Spec.ToModelJSON(ctx)
		if sErr != nil {
			err = sErr
			return
		}
		jsonProfile.Spec = &spec
	}

	if s.Interface != nil {
		ifc, iErr := s.Interface.ToModelJSON(ctx)
		if iErr != nil {
			err = iErr
			return
		}
		jsonProfile.Interface = &ifc
	}

	return
}

func (sm *SensorModel) FillFromResp(ctx context.Context, r *client.EntSensor) (err error) {
	mc := &ModelCommon{}
	helper.ConvertJSONStructToSimpleTF(ctx, r, mc)
	sm.ModelCommon = *mc

	sc := &SensorCommon{}
	err = sc.FillFromResp(ctx, r.Kind, r.Spec, r.Interface)
	if err != nil {
		return
	}
	sm.SensorCommon = *sc

	if r.Inactive != nil {
		sm.Inactive = types.BoolValue(*r.Inactive)
	}

	if r.Edges != nil && r.Edges.Profile != nil && r.Edges.Profile.Id != nil {
		sm.ProfileID = types.StringValue(*r.Edges.Profile.Id)
	}

	return
}

func (sm *SensorModel) ToModelJSON(ctx context.Context) (jsonSensor client.EntSensor, err error) {
	err = helper.ConvertTfModelToApiJSON(ctx, sm, &jsonSensor)
	if err != nil {
		return
	}

	if sm.Spec != nil {
		spec, sErr := sm.Spec.ToModelJSON(ctx)
		if sErr != nil {
			err = sErr
			return
		}
		jsonSensor.Spec = &spec
	}

	if sm.Interface != nil {
		ifc, iErr := sm.Interface.ToModelJSON(ctx)
		if iErr != nil {
			err = iErr
			return
		}
		jsonSensor.Interface = &ifc
	}

	if !sm.ProfileID.IsNull() && !sm.ProfileID.IsUnknown() {
		id := sm.ProfileID.ValueString()
		jsonSensor.Edges = &client.EntSensorEdges{
			Profile: &client.EntSensorProfile{Id: &id},
		}
	}

	return
}

func (sc *SensorCommon) FillFromResp(ctx context.Context, kind *string, rSpec *client.SchemaSensorSpec, ifc *client.SchemaSensorInterface) (err error) {
	if kind != nil {
		sc.Kind = types.StringValue(*kind)
	}

	// parse spec using its own method
	if rSpec != nil {
		spec := &SensorSpec{}
		err = spec.FillFromResp(ctx, *rSpec)
		if err != nil {
			return
		}
		sc.Spec = spec
	}

	if ifc != nil {
		interfaceSpec := &SensorInterface{}
		err = interfaceSpec.FillFromResp(ctx, *ifc)
		if err != nil {
			return
		}
		sc.Interface = interfaceSpec
	}

	return nil
}

func (s *SensorSpec) FillFromResp(ctx context.Context, resp client.SchemaSensorSpec) (err error) {
	// Handles scalar fields: HFov, MaxRange, MaxRateInHz, MinRange, MinRateInHz, RangeUnit, Units, VFoV
	helper.ConvertJSONStructToSimpleTF(ctx, resp, s)

	// map[string]string fields
	s.Comm, err = helper.ToTfStringMap(ctx, resp.Comm)
	if err != nil {
		return fmt.Errorf("comm: %w", err)
	}

	s.Media, err = helper.ToTfStringMap(ctx, resp.Media)
	if err != nil {
		return fmt.Errorf("media: %w", err)
	}

	s.Model, err = helper.ToTfStringMap(ctx, resp.Model)
	if err != nil {
		return fmt.Errorf("model: %w", err)
	}

	// map[string]float32 fields — convert to float32 map for TF
	s.Operating, err = helper.TfMapFromGoMapFloat32(ctx, resp.Operating)
	if err != nil {
		return fmt.Errorf("operating: %w", err)
	}

	s.Power, err = helper.TfMapFromGoMapFloat32(ctx, resp.Power)
	if err != nil {
		return fmt.Errorf("operating: %w", err)
	}

	if resp.Misc != nil {
		s.Misc, err = helper.JSONObjToNormalized(*resp.Misc)
		if err != nil {
			return
		}
	} else {
		s.Misc = jsontypes.NewNormalizedNull()
	}

	return nil
}

func (s *SensorSpec) ToModelJSON(ctx context.Context) (spec client.SchemaSensorSpec, err error) {
	err = helper.ConvertTfModelToApiJSON(ctx, *s, &spec)

	// map[string]string fields
	if !s.Comm.IsNull() && !s.Comm.IsUnknown() {
		comm, ok := helper.TfMapStrToGoMap(ctx, s.Comm)
		if !ok {
			err = fmt.Errorf("failed to convert comm map")
			return
		}
		spec.Comm = &comm
	}

	if !s.Media.IsNull() && !s.Media.IsUnknown() {
		media, ok := helper.TfMapStrToGoMap(ctx, s.Media)
		if !ok {
			err = fmt.Errorf("failed to convert comm map")
			return
		}
		spec.Media = &media
	}

	if !s.Model.IsNull() && !s.Model.IsUnknown() {
		model, ok := helper.TfMapStrToGoMap(ctx, s.Model)
		if !ok {
			err = fmt.Errorf("failed to convert comm map")
			return
		}
		spec.Model = &model
	}

	if !s.Operating.IsNull() && !s.Operating.IsUnknown() {
		operating, ok := helper.GoMapFloat32FromTfMap(ctx, s.Operating)
		if !ok {
			err = fmt.Errorf("failed to convert comm map")
			return
		}
		spec.Operating = &operating
	}

	if !s.Power.IsNull() && !s.Power.IsUnknown() {
		pw, ok := helper.GoMapFloat32FromTfMap(ctx, s.Operating)
		if !ok {
			err = fmt.Errorf("failed to convert comm map")
			return
		}
		spec.Power = &pw
	}

	if !s.Misc.IsNull() && !s.Misc.IsUnknown() {
		misc, iErr := helper.TfJSONToGoMapInterface(ctx, s.Misc)
		err = iErr
		if err != nil {
			err = fmt.Errorf("failed to convert comm map")
			return
		}
		spec.Misc = &misc
	}

	return
}

func (s *SensorInterface) FillFromResp(ctx context.Context, resp client.SchemaSensorInterface) (err error) {
	// Handles scalar field: Kind
	helper.ConvertJSONStructToSimpleTF(ctx, resp, s)

	s.Cable, err = helper.ToTfStringMap(ctx, resp.Cable)
	if err != nil {
		return fmt.Errorf("cable: %w", err)
	}

	s.DeviceIo, err = helper.ToTfStringMap(ctx, resp.DeviceIo)
	if err != nil {
		return fmt.Errorf("device_io: %w", err)
	}

	s.SensorIo, err = helper.ToTfStringMap(ctx, resp.SensorIo)
	if err != nil {
		return fmt.Errorf("sensor_io: %w", err)
	}

	return nil
}

func sensorSpecAttrs() []BaseSchema {
	return []BaseSchema{
		{Name: "comm", AttrType: TfMap, SubType: TfString, Optional: true, Desc: "communication parameters"},
		{Name: "h_fov", AttrType: TFFloat, Optional: true, Desc: "horizontal field of view"},
		{Name: "max_range", AttrType: TFFloat, Optional: true, Desc: "maximum range"},
		{Name: "max_rate_in_hz", AttrType: TFFloat, Optional: true, Desc: "maximum rate in Hz"},
		{Name: "media", AttrType: TfMap, SubType: TfString, Optional: true, Desc: "media parameters"},
		{Name: "min_range", AttrType: TFFloat, Optional: true, Desc: "minimum range"},
		{Name: "min_rate_in_hz", AttrType: TFFloat, Optional: true, Desc: "minimum rate in Hz"},
		{Name: "misc", AttrType: TfJSON, Optional: true, Desc: "miscellaneous parameters"},
		{Name: "model", AttrType: TfMap, SubType: TfString, Optional: true, Desc: "model parameters"},
		{Name: "operating", AttrType: TfMap, SubType: TFFloat, Optional: true, Desc: "operating parameters"},
		{Name: "power", AttrType: TfMap, SubType: TFFloat, Optional: true, Desc: "power parameters"},
		{Name: "range_unit", AttrType: TfString, Optional: true, Desc: "unit for range values"},
		{Name: "units", AttrType: TfString, Optional: true, Desc: "units"},
		{Name: "v_fov", AttrType: TFFloat, Optional: true, Desc: "vertical field of view"},
	}
}

func sensorInterfaceAttrs() []BaseSchema {
	return []BaseSchema{
		{Name: "cable", AttrType: TfMap, SubType: TfString, Optional: true, Desc: "cable interface parameters"},
		{Name: "device_io", AttrType: TfMap, SubType: TfString, Optional: true, Desc: "device IO parameters"},
		{Name: "kind", AttrType: TfString, Required: true, Desc: "interface kind"},
		{Name: "sensor_io", AttrType: TfMap, SubType: TfString, Optional: true, Desc: "sensor IO parameters"},
	}
}

func sensorAttrs() []BaseSchema {
	attrs := giveCommonAttributes()
	sensorSpecific := []BaseSchema{
		{Name: "kind", AttrType: TfString, Required: true, Desc: "sensor kind"},
		{Name: "profile_id", AttrType: TfString, Optional: true, Desc: "id of sensor profile"},
		{Name: "inactive", AttrType: TfBoolean, Optional: true, Desc: "whether the sensor is inactive"},
	}
	return append(attrs, sensorSpecific...)
}

func SensorProfileResourceSchema() rschema.Schema {
	attrs := ResAttributes(sensorProfileAttrs())
	attrs["spec"] = rschema.SingleNestedAttribute{
		Attributes:  ResAttributes(sensorSpecAttrs()),
		Required:    true,
		Description: "sensor specification",
	}
	attrs["interface"] = rschema.SingleNestedAttribute{
		Attributes:  ResAttributes(sensorInterfaceAttrs()),
		Optional:    true,
		Description: "sensor interface",
	}
	return rschema.Schema{
		Attributes:          attrs,
		Description:         "Manages a sensor profile resource.",
		MarkdownDescription: "Manages a sensor profile resource.",
	}
}

func SensorProfileDSchema() dschema.Schema {
	attrs := DSAttributes(sensorProfileAttrs())
	attrs["spec"] = dschema.SingleNestedAttribute{
		Attributes:  DSAttributes(sensorSpecAttrs()),
		Computed:    true,
		Description: "sensor specification",
	}
	attrs["interface"] = dschema.SingleNestedAttribute{
		Attributes:  DSAttributes(sensorInterfaceAttrs()),
		Computed:    true,
		Description: "sensor interface",
	}
	return dschema.Schema{
		Attributes:          attrs,
		Description:         "Fetches a sensor profile by ID.",
		MarkdownDescription: "Fetches a sensor profile by ID.",
	}
}

func sensorProfileAttrs() []BaseSchema {
	attrs := giveCommonAttributes()
	sensorProfileSpecific := []BaseSchema{
		{Name: "kind", AttrType: TfString, Optional: true, Desc: "sensor kind"},
	}
	return append(attrs, sensorProfileSpecific...)
}

func SensorResourceSchema() rschema.Schema {
	attrs := ResAttributes(sensorAttrs())
	attrs["spec"] = rschema.SingleNestedAttribute{
		Attributes:  ResAttributes(sensorSpecAttrs()),
		Optional:    true,
		Description: "sensor specification",
	}
	attrs["interface"] = rschema.SingleNestedAttribute{
		Attributes:  ResAttributes(sensorInterfaceAttrs()),
		Optional:    true,
		Description: "sensor interface",
	}
	return rschema.Schema{
		Attributes:          attrs,
		Description:         "Manages a sensor resource.",
		MarkdownDescription: "Manages a sensor resource.",
	}
}

func SensorDSchema() dschema.Schema {
	attrs := DSAttributes(sensorAttrs())
	attrs["spec"] = dschema.SingleNestedAttribute{
		Attributes:  DSAttributes(sensorSpecAttrs()),
		Computed:    true,
		Description: "sensor specification",
	}
	attrs["interface"] = dschema.SingleNestedAttribute{
		Attributes:  DSAttributes(sensorInterfaceAttrs()),
		Computed:    true,
		Description: "sensor interface",
	}
	return dschema.Schema{
		Attributes:          attrs,
		Description:         "Fetches a sensor by ID.",
		MarkdownDescription: "Fetches a sensor by ID.",
	}
}

func (s *SensorInterface) ToModelJSON(ctx context.Context) (iface client.SchemaSensorInterface, err error) {
	err = helper.ConvertTfModelToApiJSON(ctx, *s, &iface)

	if !s.Cable.IsNull() && !s.Cable.IsUnknown() {
		cable, ok := helper.TfMapStrToGoMap(ctx, s.Cable)
		if !ok {
			err = fmt.Errorf("failed to convert cable map")
			return
		}
		iface.Cable = &cable
	}

	if !s.DeviceIo.IsNull() && !s.DeviceIo.IsUnknown() {
		deviceIo, ok := helper.TfMapStrToGoMap(ctx, s.DeviceIo)
		if !ok {
			err = fmt.Errorf("failed to convert device_io map")
			return
		}
		iface.DeviceIo = &deviceIo
	}

	if !s.SensorIo.IsNull() && !s.SensorIo.IsUnknown() {
		sensorIo, ok := helper.TfMapStrToGoMap(ctx, s.SensorIo)
		if !ok {
			err = fmt.Errorf("failed to convert sensor_io map")
			return
		}
		iface.SensorIo = &sensorIo
	}

	return
}
