// Copyright (c) ArenaML Labs Pvt Ltd.

package schema

import (
	"context"
	"sort"

	"github.com/arena-ml/terraform-provider-arenaml/generator/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type ModelCommon struct {
	ID          types.String `tfsdk:"id" json:"id"`
	Name        types.String `tfsdk:"name" json:"name"`
	Description types.String `tfsdk:"description" json:"description"`
	Created     types.String `tfsdk:"created" json:"created"`
	Updated     types.String `tfsdk:"updated" json:"updated"`
	Version     types.Int64  `tfsdk:"version" json:"version"`
}

func giveCommonAttributes() []BaseSchema {
	return []BaseSchema{
		{
			Name:     "id",
			Optional: true,
			Computed: true,
			AttrType: TfString,
			Desc:     "The unique identifier for the basis",
		},
		{
			Name:     "description",
			Optional: true,
			AttrType: TfString,
			Desc:     "A description of the basis",
		},
		{
			Name:     "name",
			AttrType: TfString,
			Desc:     "The unique name of the basis",
			Required: true,
		},
		{
			Name:     "created",
			Computed: true,
			AttrType: TfString,
			Desc:     "The timestamp when the basis was created",
		},
		{
			Name:     "updated",
			Computed: true,
			AttrType: TfString,
			Desc:     "The timestamp when the basis was last updated",
		},
		{
			Name:     "version",
			Computed: true,
			AttrType: TfInt64,
			Desc:     "The version of the basis",
		},
	}
}

type Tag struct {
	Key   basetypes.StringValue `tfsdk:"key"`
	Value basetypes.StringValue `tfsdk:"value"`
}

func ConvertTags(ctx context.Context, tagResp []client.TagsTag) []Tag {
	if tagResp == nil || len(tagResp) == 0 {
		return []Tag{}
	}

	tagList := make([]Tag, len(tagResp))

	sort.Slice(tagResp, func(i, j int) bool {
		if tagResp[i].Key == nil {
			return false
		}
		if tagResp[j].Key == nil {
			return true
		}

		return *tagResp[i].Key > *tagResp[j].Key
	})

	for i, tagResp := range tagResp {
		tag := Tag{}
		if tagResp.Key != nil {
			tag.Key = types.StringValue(*tagResp.Key)
		}
		if tagResp.Value != nil {
			tag.Value = types.StringValue(*tagResp.Value)
		}

		tagList[i] = tag
	}

	return tagList
}
