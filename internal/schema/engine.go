// Copyright (c) ArenaML Labs Pvt Ltd.

package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EngineModel struct {
	Id       types.String         `tfsdk:"id"`
	Inactive types.Bool           `tfsdk:"inactive"`
	Kind     types.String         `tfsdk:"kind"`
	Name     types.String         `tfsdk:"name"`
	Spec     jsontypes.Normalized `tfsdk:"spec"`
	Tags     types.List           `tfsdk:"tags"`
}

func EngineDataSourceSchema(ctx context.Context) dschema.Schema {
	return dschema.Schema{
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The unique identifier for the engine",
				MarkdownDescription: "The unique identifier for the engine",
			},
			"name": dschema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The unique name of the engine",
				MarkdownDescription: "The unique name of the engine",
			},
			"inactive": dschema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the engine is inactive",
				MarkdownDescription: "Whether the engine is inactive",
			},
			"kind": dschema.StringAttribute{
				Computed:            true,
				Description:         "The kind of engine",
				MarkdownDescription: "The kind of engine",
			},
			"spec": dschema.StringAttribute{
				CustomType:          jsontypes.NormalizedType{},
				Computed:            true,
				Description:         "Engine-specific configuration details",
				MarkdownDescription: "Engine-specific configuration details",
			},
			"tags": dschema.ListNestedAttribute{
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"key": dschema.StringAttribute{
							Computed: true,
						},
						"value": dschema.StringAttribute{
							Computed: true,
						},
					},
				},
				Computed:            true,
				Description:         "Tags associated with the engine",
				MarkdownDescription: "Tags associated with the engine",
			},
		},
	}
}

func EngineResourceSchema(ctx context.Context) rschema.Schema {
	return rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier for the engine",
				MarkdownDescription: "The unique identifier for the engine",
			},
			"inactive": rschema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether the engine is inactive",
				MarkdownDescription: "Whether the engine is inactive",
			},
			"kind": rschema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString("nomad"),
				Description:         "The kind of engine",
				MarkdownDescription: "The kind of engine",
				Validators: []validator.String{
					stringvalidator.OneOf("nomad"),
				},
			},
			"name": rschema.StringAttribute{
				Required:            true,
				Description:         "The unique name of the engine",
				MarkdownDescription: "The unique name of the engine",
			},
			"spec": rschema.StringAttribute{
				Required:            true,
				CustomType:          jsontypes.NormalizedType{},
				Description:         "Engine-specific configuration details",
				MarkdownDescription: "Engine-specific configuration details",
			},
			"tags": rschema.ListNestedAttribute{
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"key": rschema.StringAttribute{
							Computed: true,
						},
						"value": rschema.StringAttribute{
							Computed: true,
						},
					},
				},
				Optional:            true,
				Description:         "Tags associated with the engine",
				MarkdownDescription: "Tags associated with the engine",
			},
		},
	}
}
