// Copyright (c) ArenaML Labs Pvt Ltd.

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
)

var s client.EntStorage

type Store struct {
	ModelCommon
	Basepath types.String `tfsdk:"basepath"`
	Disabled types.Bool   `tfsdk:"disabled"`
	Endpoint types.String `tfsdk:"endpoint"`
	Kind     types.String `tfsdk:"kind"`
	OrgId    types.String `tfsdk:"org_id"`
	ReadOnly types.Bool   `tfsdk:"read_only"`
	Config   *StoreConfig `tfsdk:"config"`
}

type StoreConfig struct {
	Auth            jsontypes.Normalized `tfsdk:"auth"`
	Spec            jsontypes.Normalized `tfsdk:"spec"`
	MaxObjects      types.Int64          `tfsdk:"max_objects"`
	CapacityGB      types.Int64          `tfsdk:"capacity_gb"`
	MaxObjectSizeMB types.Int64          `tfsdk:"max_object_size_mb"`
}

func (c *StoreConfig) FillFromResp(ctx context.Context, resp client.EntStorage) (err error) {
	if resp.Config == nil {
		return nil
	}

	conf := *resp.Config
	helper.ConvertJSONStructToSimpleTF(ctx, conf, c)

	if conf.Auth != nil {
		c.Auth, err = helper.JSONObjToNormalized(*conf.Auth)
	} else {
		c.Auth = jsontypes.NewNormalizedNull()
	}

	if conf.Spec != nil {
		c.Spec, err = helper.JSONObjToNormalized(*conf.Spec)
	} else {
		c.Spec = jsontypes.NewNormalizedNull()
	}

	helper.ConvertJSONStructToSimpleTF(ctx, conf, c)

	return err
}

func (c *StoreConfig) ToModelJson(ctx context.Context) (jsonConf client.SchemaStoreConfig, err error) {
	err = helper.ConvertTfModelToApiJSON(ctx, *c, &jsonConf)

	if !c.Auth.IsNull() && !c.Auth.IsUnknown() {
		auth, err := helper.TfJSONToGoMapInterface(ctx, c.Auth)
		if err != nil {
			err = fmt.Errorf("auth not found in tf data: \n %s %s ", c.Auth, err)
			return jsonConf, err
		}
		jsonConf.Auth = &auth
	}

	if !c.Spec.IsNull() && !c.Spec.IsUnknown() {
		spec, err := helper.TfJSONToGoMapInterface(ctx, c.Spec)
		if err != nil {
			err = fmt.Errorf("spec not found in tf data: \n %s %s ", c.Spec, err)
			return jsonConf, err
		}
		jsonConf.Spec = &spec
	}

	return
}

func (s *Store) FillFromResp(ctx context.Context, resp client.EntStorage) (err error) {
	helper.ConvertJSONStructToSimpleTF(ctx, resp, s)
	mc := &ModelCommon{}
	helper.ConvertJSONStructToSimpleTF(ctx, resp, mc)
	s.ModelCommon = *mc

	if resp.Config != nil {
		storeConf := &StoreConfig{}
		err = storeConf.FillFromResp(ctx, resp)
		if err != nil {
			return
		}
		s.Config = storeConf
	}

	return nil
}

func (s *Store) ToModelJSON(ctx context.Context) (jsonStore client.EntStorage, err error) {
	err = helper.ConvertTfModelToApiJSON(ctx, s.ModelCommon, &jsonStore)
	if err != nil {
		return
	}
	err = helper.ConvertTfModelToApiJSON(ctx, s, &jsonStore)

	var clStoreConf client.SchemaStoreConfig
	clStoreConf, err = s.Config.ToModelJson(ctx)
	if err != nil {
		return
	}
	jsonStore.Config = &clStoreConf

	return
}

func storeConfigAttrs() []BaseSchema {
	return []BaseSchema{
		{
			Name:      "auth",
			AttrType:  TfJSON,
			Optional:  true,
			Sensitive: true,
			Desc:      "authentication configuration for the store",
		},
		{
			Name:     "spec",
			AttrType: TfJSON,
			Optional: true,
			Desc:     "storage specification configuration",
		},
		{
			Name:     "max_objects",
			AttrType: TfInt64,
			Optional: true,
			Desc:     "maximum number of objects",
		},
		{
			Name:     "capacity_gb",
			AttrType: TfInt64,
			Optional: true,
			Desc:     "storage capacity in GB",
		},
		{
			Name:     "max_object_size_mb",
			AttrType: TfInt64,
			Optional: true,
			Desc:     "maximum object size in MB",
		},
	}
}

const storeConfigAttrDesc = "configuration for the store"

func dsStoreConfigSchema() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		Attributes:          DSAttributes(storeConfigAttrs()),
		Computed:            true,
		Description:         storeConfigAttrDesc,
		MarkdownDescription: storeConfigAttrDesc,
	}
}

func resStoreConfigSchema() rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Attributes:          ResAttributes(storeConfigAttrs()),
		Required:            true,
		Description:         storeConfigAttrDesc,
		MarkdownDescription: storeConfigAttrDesc,
	}
}

func storeAttrs() []BaseSchema {
	commonAttrs := giveCommonAttributes()
	attrs := []BaseSchema{
		{
			Name:     "basepath",
			AttrType: TfString,
			Optional: true,
			Desc:     "base path for the store",
		},
		{
			Name:     "disabled",
			AttrType: TfBoolean,
			Optional: true,
			Desc:     "whether the store is disabled",
		},
		{
			Name:     "endpoint",
			AttrType: TfString,
			Required: true,
			Desc:     "endpoint URL for the store",
		},
		{
			Name:     "kind",
			AttrType: TfString,
			Required: true,
			Desc:     "kind of storage backend",
		},
		{
			Name:     "org_id",
			AttrType: TfString,
			Required: true,
			Desc:     "organization id this store belongs to",
		},
		{
			Name:     "read_only",
			AttrType: TfBoolean,
			Optional: true,
			Desc:     "whether the store is read-only",
		},
	}

	attrs = append(commonAttrs, attrs...)

	return attrs
}

func StoreDSchema() dschema.Schema {
	attrs := DSAttributes(storeAttrs())
	attrs["config"] = dsStoreConfigSchema()

	return dschema.Schema{
		Attributes:  attrs,
		Description: "storage resource",
	}
}

func StoreResourceSchema() rschema.Schema {
	attrs := ResAttributes(storeAttrs())
	attrs["config"] = resStoreConfigSchema()

	return rschema.Schema{
		Attributes:  attrs,
		Description: "storage resource",
	}
}
