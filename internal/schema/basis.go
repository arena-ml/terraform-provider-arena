// Copyright (c) ArenaML Labs Pvt Ltd.

package schema

import (
	"context"
	"fmt"

	"github.com/arena-ml/terraform-provider-arenaml/generator/client"
	"github.com/arena-ml/terraform-provider-arenaml/helper"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type BasisSource struct {
	Format types.String `tfsdk:"format" json:"format"`
	Raw    types.String `tfsdk:"raw" json:"raw"`
}

func (s *BasisSource) FillFromResp(ctx context.Context, resp client.EntBasis) error {
	if resp.Config == nil || resp.Config.Source == nil {
		return fmt.Errorf("basis source in config is nill")
	}

	respSource := resp.Config.Source
	helper.ConvertJSONStructToSimpleTF(ctx, *respSource, s)

	return nil
}

type BasisCollector struct {
	BounceDuration types.String         `tfsdk:"bounce_duration" json:"bounce_duration"`
	MaxParallel    types.Int64          `tfsdk:"max_parallel" json:"max_parallel"`
	MaxSize        types.String         `tfsdk:"max_size" json:"max_size"`
	RunSpec        jsontypes.Normalized `tfsdk:"run_spec" json:"run_spec"`
	Timeout        types.String         `tfsdk:"timeout" json:"timeout"`
}

func (c *BasisCollector) FillFromResp(ctx context.Context, resp client.EntBasis) (err error) {
	if resp.Config == nil || resp.Config.Collector == nil {
		tflog.Info(ctx, "basis collector in config is nill")
		return
	}

	respCollector := resp.Config.Collector
	helper.ConvertJSONStructToSimpleTF(ctx, *respCollector, c)

	c.RunSpec, err = helper.JSONObjToNormalized(respCollector.RunSpec)

	return
}

type BasisWatcher struct {
	Env          types.Map            `tfsdk:"env" json:"env"`
	Image        types.String         `tfsdk:"image" json:"image"`
	NoCollect    types.Bool           `tfsdk:"no_collect" json:"no_collect"`
	RegistryAuth types.Map            `tfsdk:"registry_auth" json:"registry_auth"`
	RunSpec      jsontypes.Normalized `tfsdk:"run_spec" json:"run_spec"`
}

func (w *BasisWatcher) FillFromResp(ctx context.Context, resp client.EntBasis) (err error) {
	if resp.Config == nil || resp.Config.Watcher == nil {
		return fmt.Errorf("required field watcher in basis config is ni:w" +
			"l")
	}

	respWatcher := resp.Config.Watcher
	helper.ConvertJSONStructToSimpleTF(ctx, *respWatcher, w)

	w.RegistryAuth, err = helper.ToTfStringMap(ctx, respWatcher.RegistryAuth)
	if err != nil {
		return err
	}

	w.Env, err = helper.ToTfStringMap(ctx, respWatcher.Env)
	if err != nil {
		return err
	}

	if respWatcher.RunSpec != nil {
		w.RunSpec, err = helper.JSONObjToNormalized(respWatcher.RunSpec)
	}
	return nil
}

type BasisModel struct {
	AllowedEnvs types.List      `tfsdk:"allowed_envs" json:"allowed_envs"`
	Source      *BasisSource    `tfsdk:"source" json:"source,omitempty"`
	Watcher     *BasisWatcher   `tfsdk:"watcher" json:"watcher,omitempty"`
	Collector   *BasisCollector `tfsdk:"collector" json:"collector,omitempty"`
	Frozen      types.Bool      `tfsdk:"frozen" json:"frozen"`
	Kind        types.String    `tfsdk:"kind" json:"kind"`
	ModelCommon
}

func (b *BasisModel) FillFromResp(ctx context.Context, resp client.EntBasis) error {
	b.Source = &BasisSource{}
	b.Watcher = &BasisWatcher{}

	helper.ConvertJSONStructToSimpleTF(ctx, resp, b)
	m := &ModelCommon{}
	helper.ConvertJSONStructToSimpleTF(ctx, resp, m)
	b.ModelCommon = *m

	if err := b.Source.FillFromResp(ctx, resp); err != nil {
		return err
	}
	if err := b.Watcher.FillFromResp(ctx, resp); err != nil {
		return err
	}

	if resp.Config.Collector != nil {
		b.Collector = &BasisCollector{}
		if err := b.Collector.FillFromResp(ctx, resp); err != nil {
			return err
		}
	}

	var diags diag.Diagnostics
	if resp.Config == nil || resp.Config.AllowedEnvs == nil {
		b.AllowedEnvs, diags = types.ListValueFrom(ctx, types.StringType, resp.Config.AllowedEnvs)
	} else {
		b.AllowedEnvs = types.ListNull(types.StringType)
	}

	if diags.HasError() {
		return fmt.Errorf(diags[0].Summary())
	}

	return nil
}

// BasisDataSourceAttrs returns the attribute definitions for the top-level Basis data source schema
// excluding the nested watcher, collector, and source blocks which are added separately.
func BasisDataSourceAttrs() []BaseSchema {
	commonAttrs := giveCommonAttributes()
	basisAttrs := []BaseSchema{
		{
			Name:     "allowed_envs",
			AttrType: TFList,
			Optional: true,
			SubType:  TfString,
			Desc:     "limit the execution of basis watcher & colletor to these env, if empty then only primary env is allowed",
		},
		{
			Name:     "frozen",
			Optional: true,
			AttrType: TfBoolean,
			Desc:     "Whether the basis is frozen",
		},
		{
			Name:     "kind",
			Required: true,
			AttrType: TfString,
			Desc:     "The kind of basis (e.g., git, s3, gcs, etc.)",
		},
	}

	return append(commonAttrs, basisAttrs...)
}

func BasisDataSourceSchema(_ context.Context) dschema.Schema {
	attrs := DSAttributes(BasisDataSourceAttrs())
	// Set nested attributes after generating map from BaseSchema list
	attrs["watcher"] = dsBasisWatcherSchema()
	attrs["collector"] = dsBasisCollectorSchema()
	attrs["source"] = dsBasisSourceSchema()

	return dschema.Schema{
		Attributes: attrs,
	}
}

// BasisWatcherAttrs returns the attribute definitions for the Basis watcher nested schema.
func BasisWatcherAttrs() []BaseSchema {
	return []BaseSchema{
		{
			Name:     "env",
			AttrType: TfMap,
			SubType:  TfString,
			Optional: true,
			Desc:     "env variable to set for the container image",
		},
		{
			Name:     "image",
			Required: true,
			AttrType: TfString,
			Desc:     "container image name",
		},
		{
			Name:     "no_collect",
			Optional: true,
			Computed: true,
			AttrType: TfBoolean,
			Desc:     "if true then a separate collect job will launch to get new version. To handle cases where collection can take a lot of resources and time",
		},
		{
			Name:      "registry_auth",
			Optional:  true,
			AttrType:  TfMap,
			SubType:   TfString,
			Sensitive: false,
			Desc:      "auth for registry if image is not public",
		},
		{
			Name:     "run_spec",
			Optional: true,
			AttrType: TfJSON,
			Desc:     "resource and other limits along with driver override, other than docker only podman can be used for now",
		},
	}
}

const basisWatcherAttrDesc = "( required ) watcher image and other config details, for now collector uses the same image as watcher"

func dsBasisWatcherSchema() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		Attributes:          DSAttributes(BasisWatcherAttrs()),
		Computed:            true,
		Description:         basisWatcherAttrDesc,
		MarkdownDescription: basisWatcherAttrDesc,
	}
}

// BasisCollectorAttrs returns the attribute definitions for the Basis collector nested schema.
func BasisCollectorAttrs() []BaseSchema {
	return []BaseSchema{
		{
			Name:     "bounce_duration",
			Optional: true,
			AttrType: TfString,
			Desc:     "format time.ParseDuration e.g. 1h, 30m, in case not all values needs to be collected, for e.g.",
		},
		{
			Name:     "max_parallel",
			Optional: true,
			AttrType: TfInt64,
			Desc:     "max number of concurrent collect task to run",
		},
		{
			Name:     "max_size",
			Optional: true,
			AttrType: TfString,
			Desc:     "max size of collected artifact in datasize.ParseString format eg. 1kb, 500mb",
		},
		{
			Name:     "run_spec",
			Optional: true,
			AttrType: TfJSON,
			Desc:     "misc compute resource details",
		},
		{
			Name:     "timeout",
			Optional: true,
			AttrType: TfString,
			Desc:     "format time.ParseDuration e.g. 1h, 30m",
		},
	}
}

const basisCollectorAttrDesc = "collector config when artifact are fetched by a different process instead of watcher"

func dsBasisCollectorSchema() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		Attributes:  DSAttributes(BasisCollectorAttrs()),
		Computed:    true,
		Description: basisCollectorAttrDesc,
	}
}

// BasisSourceAttrs returns the attribute definitions for the Basis source nested schema.
// This separates the attribute list from the schema construction to enable reuse
// and keep dsBasisSourceSchema concise.
func BasisSourceAttrs() []BaseSchema {
	return []BaseSchema{
		{
			Name:     "format",
			AttrType: TfString,
			Required: true,
			Desc:     "if empty then assumed json",
		},
		{
			Name:     "raw",
			Required: true,
			AttrType: TfString,
			Desc:     "template string used directly",
		},
	}
}

const basisSourceAttrDesc = "( required ) source config of basis to be used by both watcher to check for new version"

func dsBasisSourceSchema() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		Attributes:          DSAttributes(BasisSourceAttrs()),
		Computed:            true,
		Description:         basisSourceAttrDesc,
		MarkdownDescription: basisSourceAttrDesc,
	}
}

func resBasisWatcherSchema() rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Attributes:          ResAttributes(BasisWatcherAttrs()),
		Required:            true,
		Description:         basisWatcherAttrDesc,
		MarkdownDescription: basisWatcherAttrDesc,
	}
}

func resBasisSourceSchema() rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Attributes:          ResAttributes(BasisSourceAttrs()),
		Required:            true,
		Description:         basisSourceAttrDesc,
		MarkdownDescription: basisSourceAttrDesc,
	}
}

func resBasisCollectorSchema() rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Attributes:          ResAttributes(BasisCollectorAttrs()),
		Optional:            true,
		Description:         basisCollectorAttrDesc,
		MarkdownDescription: basisCollectorAttrDesc,
	}
}

func BasisResourceSchema() rschema.Schema {
	attrs := ResAttributes(BasisDataSourceAttrs())
	// Set nested attributes after generating map from BaseSchema list
	attrs["watcher"] = resBasisWatcherSchema()

	attrs["collector"] = resBasisCollectorSchema()
	attrs["source"] = resBasisSourceSchema()

	return rschema.Schema{
		Attributes: attrs,
	}
}
