// Copyright (c) ArenaML Labs Pvt Ltd.

package schema

import (
	"context"
	"fmt"

	"github.com/arena-ml/terraform-provider-arenaml/fixed"
	"github.com/arena-ml/terraform-provider-arenaml/generator/client"
	"github.com/arena-ml/terraform-provider-arenaml/helper"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kr/pretty"
)

type NodeStep struct {
	PipelineID types.String `tfsdk:"pipeline_id"`
	ModelCommon
	Kind     types.String `tfsdk:"kind"`
	StoreID  types.String `tfsdk:"store_id"`
	EngineID types.String `tfsdk:"engine_id"`
	Config   *StepConfig  `tfsdk:"config"`
}

func (ns *NodeStep) Verify() error {
	conf := ns.Config
	kind := ns.Kind.ValueString()
	if kind == fixed.StepKindContainer &&
		(conf.Image.IsNull() || conf.Image.IsUnknown() || conf.Image.String() == "") {
		return fmt.Errorf("step config's image field is required for docker kind")
	}

	if (kind == fixed.StepKindExec || kind == fixed.StepKindRawExec) &&
		(conf.Command.IsNull() || conf.Command.IsUnknown() || conf.Command.String() == "") {
		return fmt.Errorf("step config's command field is required for %s or %s kind", fixed.StepKindExec, fixed.StepKindRawExec)
	}

	return nil
}

type StepConfig struct {
	Image        types.String         `tfsdk:"image"`
	Privileged   types.Bool           `tfsdk:"privileged"`
	Command      types.String         `tfsdk:"command"`
	Args         types.List           `tfsdk:"args"`
	EnvVars      types.Map            `tfsdk:"env_vars"`
	RegistryAuth types.Map            `tfsdk:"registry_auth"`
	RunSpec      jsontypes.Normalized `tfsdk:"run_spec"`
}

func (c *StepConfig) FillFromResp(ctx context.Context, resp client.EntStep) (err error) {
	tflog.Warn(ctx, fmt.Sprintf("\n\n%s\n\n", pretty.Sprint(c)))
	if resp.Config == nil {
		return nil
	}

	// the struct value needs to be based to helper for reflection to tf model
	conf := *resp.Config
	helper.ConvertJSONStructToSimpleTF(ctx, conf, c)

	if conf.RegistryAuth != nil {
		c.RegistryAuth, err = helper.ToTfStringMap(ctx, conf.RegistryAuth)
		if err != nil {
			return
		}
	} else {
		c.RegistryAuth = basetypes.NewMapNull(basetypes.StringType{})
	}

	if conf.Args != nil {
		c.Args, err = helper.FromGoStrSliceToTfList(ctx, *conf.Args)
		if err != nil {
			return
		}
	} else {
		c.Args = basetypes.NewListNull(basetypes.StringType{})
	}

	if conf.EnvVars != nil {
		c.EnvVars, err = helper.ToTfStringMap(ctx, conf.EnvVars)
		if err != nil {
			return err
		}
	} else {
		c.EnvVars = basetypes.NewMapNull(basetypes.StringType{})
	}

	if conf.RunSpec != nil {
		c.RunSpec, err = helper.JSONObjToNormalized(*conf.RunSpec)
	} else {
		c.RunSpec = jsontypes.NewNormalizedNull()
	}

	tflog.Warn(ctx, fmt.Sprintf("\n to JSON\n%s\n\n", pretty.Sprint(c, resp.Config)))
	return err
}

func (c *StepConfig) ToModelJson(ctx context.Context) (jsonConf client.SchemaStepConfig, err error) {
	err = helper.ConvertTfModelToApiJSON(ctx, *c, &jsonConf)

	env, ok := helper.TfMapStrToGoMap(ctx, c.EnvVars)
	if !ok {
		err = fmt.Errorf("env var not found in tf env")
		return
	}
	jsonConf.EnvVars = &env

	if !c.RunSpec.IsNull() && !c.RunSpec.IsUnknown() {
		runSpec, err := helper.TfJSONToGoMapInterface(ctx, c.RunSpec)
		if err != nil {
			err = fmt.Errorf("run spec not found in tf data: \n %s %s ", c.RunSpec, err)
			return jsonConf, err
		}
		jsonConf.RunSpec = &runSpec
	}

	if !c.Args.IsNull() && !c.Args.IsUnknown() {
		args, ok := helper.TfListStrToGoSlice(ctx, c.Args)
		if !ok {
			err = fmt.Errorf("args not found in tf args")
			return
		}
		jsonConf.Args = &args
	}

	if !c.RegistryAuth.IsNull() && !c.RegistryAuth.IsUnknown() {
		regAuth, ok := helper.TfMapStrToGoMap(ctx, c.RegistryAuth)
		if !ok {
			err = fmt.Errorf("regAuth not found in tf registry_auth")
			return
		}
		jsonConf.RegistryAuth = &regAuth
	}

	tflog.Warn(ctx, fmt.Sprintf("\nto TF\n%s\n\n", pretty.Sprint(c, jsonConf)))
	return
}

func (ns *NodeStep) FillFromResp(ctx context.Context, resp client.EntStep) (err error) {
	helper.ConvertJSONStructToSimpleTF(ctx, resp, ns)
	mc := &ModelCommon{}
	helper.ConvertJSONStructToSimpleTF(ctx, resp, mc)
	ns.ModelCommon = *mc

	if resp.Config != nil {
		stepConf := &StepConfig{}
		err = stepConf.FillFromResp(ctx, resp)
		if err != nil {
			return
		}
		ns.Config = stepConf
	}

	return nil
}

func (ns *NodeStep) ToModelJSON(ctx context.Context) (jsonStep client.EntStep, err error) {
	err = helper.ConvertTfModelToApiJSON(ctx, ns.ModelCommon, &jsonStep)
	if err != nil {
		return
	}
	err = helper.ConvertTfModelToApiJSON(ctx, ns, &jsonStep)

	var clStepConf client.SchemaStepConfig
	clStepConf, err = ns.Config.ToModelJson(ctx)
	if err != nil {
		return
	}
	jsonStep.Config = &clStepConf

	return
}

func stepConfigAttrs() []BaseSchema {
	return []BaseSchema{
		{
			Name:     "image",
			AttrType: TfString,
			Optional: true,
			Desc:     "container image for the step",
		},
		{
			Name:     "privileged",
			AttrType: TfBoolean,
			Optional: true,
			Desc:     "run container in privileged mode",
		},
		{
			Name:     "command",
			AttrType: TfString,
			Optional: true,
			Desc:     "command to execute in the container",
		},
		{
			Name:     "args",
			AttrType: TFList,
			Optional: true,
			SubType:  TfString,
			Desc:     "arguments for the command",
		},
		{
			Name:     "env_vars",
			AttrType: TfMap,
			Optional: true,
			SubType:  TfString,
			Desc:     "environment variables for the container",
		},
		{
			Name:      "registry_auth",
			AttrType:  TfMap,
			Optional:  true,
			Sensitive: true,
			SubType:   TfString,
			Desc:      "authentication credentials for registry",
		},
		{
			Name:     "run_spec",
			AttrType: TfJSON,
			Required: true,
			Desc:     "resource and other limits along with driver override",
		},
	}
}

const stepConfigAttrDesc = "configuration for the step execution"

func dsStepConfigSchema() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		Attributes:          DSAttributes(stepConfigAttrs()),
		Computed:            true,
		Description:         stepConfigAttrDesc,
		MarkdownDescription: stepConfigAttrDesc,
	}
}

func resStepConfigSchema() rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Attributes:          ResAttributes(stepConfigAttrs()),
		Required:            true,
		Description:         stepConfigAttrDesc,
		MarkdownDescription: stepConfigAttrDesc,
	}
}

func nodeStepAttrs() []BaseSchema {
	attrs := giveCommonAttributes()
	stepAttrs := []BaseSchema{
		{
			Name:     "pipeline_id",
			AttrType: TfString,
			Required: true,
			Desc:     "id of the pipeline this step is part of",
		},
		{
			Name:     "kind",
			AttrType: TfString,
			Required: true,
			Desc:     "id of the pipeline this step is part of",
		},
		{
			Name:     "store_id",
			AttrType: TfString,
			Optional: true,
			Desc:     "override for store to be used in place of default store for that env",
		},
		{
			Name:     "engine_id",
			AttrType: TfString,
			Optional: true,
			Desc:     "engine to use for executing this step",
		},
	}

	stepAttrs = append(attrs, stepAttrs...)

	return stepAttrs
}

func NodeStepDSchema() dschema.Schema {
	attrs := DSAttributes(nodeStepAttrs())
	attrs["config"] = dsStepConfigSchema()

	return dschema.Schema{
		Attributes:  attrs,
		Description: "node of type step",
	}
}

func NodeStepResourceSchema() rschema.Schema {
	attrs := ResAttributes(nodeStepAttrs())
	attrs["config"] = resStepConfigSchema()

	return rschema.Schema{
		Attributes:  attrs,
		Description: "node of type step",
	}
}
