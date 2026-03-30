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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kr/pretty"
)

type User struct {
	ModelCommon
	EMail  types.String `tfsdk:"email"`
	Active types.Bool   `tfsdk:"active"`
	Config *UserConfig  `tfsdk:"config"`
}

type UserConfig struct {
	Auth   jsontypes.Normalized `tfsdk:"auth"`
	Meta   types.Map            `tfsdk:"meta"`
	Tokens jsontypes.Normalized `tfsdk:"tokens"`
}

func (c *UserConfig) FillFromResp(ctx context.Context, resp client.EntUser) (err error) {
	tflog.Warn(ctx, fmt.Sprintf("\n\n%s\n\n", pretty.Sprint(c)))
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

	if conf.Meta != nil {
		c.Meta, err = helper.ToTfStringMap(ctx, conf.Meta)
		if err != nil {
			return
		}
	} else {
		c.Meta = basetypes.NewMapNull(basetypes.StringType{})
	}

	if conf.Tokens != nil {
		tokensMap := make(map[string]interface{})
		for k, v := range *conf.Tokens {
			tokenData := map[string]interface{}{}
			if v.Name != nil {
				tokenData["name"] = *v.Name
			}
			if v.Value != nil {
				tokenData["value"] = *v.Value
			}
			if v.Expires != nil {
				tokenData["expires"] = *v.Expires
			}
			tokensMap[k] = tokenData
		}
		c.Tokens, err = helper.JSONObjToNormalized(tokensMap)
	} else {
		c.Tokens = jsontypes.NewNormalizedNull()
	}

	tflog.Warn(ctx, fmt.Sprintf("\n to JSON\n%s\n\n", pretty.Sprint(c, resp.Config)))
	return err
}

func (c *UserConfig) ToModelJson(ctx context.Context) (jsonConf client.SchemaUserConfig, err error) {
	err = helper.ConvertTfModelToApiJSON(ctx, *c, &jsonConf)

	if !c.Auth.IsNull() && !c.Auth.IsUnknown() {
		auth, err := helper.TfJSONToGoMapInterface(ctx, c.Auth)
		if err != nil {
			err = fmt.Errorf("auth not found in tf data: \n %s %s ", c.Auth, err)
			return jsonConf, err
		}
		jsonConf.Auth = &auth
	}

	if !c.Meta.IsNull() && !c.Meta.IsUnknown() {
		meta, ok := helper.TfMapStrToGoMap(ctx, c.Meta)
		if !ok {
			err = fmt.Errorf("meta not found in tf data")
			return
		}
		jsonConf.Meta = &meta
	}

	if !c.Tokens.IsNull() && !c.Tokens.IsUnknown() {
		tokensRaw, err := helper.TfJSONToGoMapInterface(ctx, c.Tokens)
		if err != nil {
			err = fmt.Errorf("tokens not found in tf data: \n %s %s ", c.Tokens, err)
			return jsonConf, err
		}

		tokens := make(map[string]client.SchemaToken)
		for k, v := range tokensRaw {
			tokenMap, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			token := client.SchemaToken{}
			if name, ok := tokenMap["name"].(string); ok {
				token.Name = &name
			}
			if value, ok := tokenMap["value"].(string); ok {
				token.Value = &value
			}
			if expires, ok := tokenMap["expires"].(string); ok {
				token.Expires = &expires
			}
			tokens[k] = token
		}
		jsonConf.Tokens = &tokens
	}

	tflog.Warn(ctx, fmt.Sprintf("\nto TF\n%s\n\n", pretty.Sprint(c, jsonConf)))
	return
}

func (u *User) FillFromResp(ctx context.Context, resp client.EntUser) (err error) {
	helper.ConvertJSONStructToSimpleTF(ctx, resp, u)
	mc := &ModelCommon{}
	helper.ConvertJSONStructToSimpleTF(ctx, resp, mc)
	u.ModelCommon = *mc

	if resp.Config != nil {
		userConf := &UserConfig{}
		err = userConf.FillFromResp(ctx, resp)
		if err != nil {
			return
		}
		u.Config = userConf
	}

	return nil
}

func (u *User) ToModelJSON(ctx context.Context) (jsonUser client.EntUser, err error) {
	err = helper.ConvertTfModelToApiJSON(ctx, u.ModelCommon, &jsonUser)
	if err != nil {
		return
	}
	err = helper.ConvertTfModelToApiJSON(ctx, u, &jsonUser)

	var clUserConf client.SchemaUserConfig
	clUserConf, err = u.Config.ToModelJson(ctx)
	if err != nil {
		return
	}
	jsonUser.Config = &clUserConf

	return
}

func userConfigAttrs() []BaseSchema {
	return []BaseSchema{
		{
			Name:      "auth",
			AttrType:  TfJSON,
			Optional:  true,
			Sensitive: true,
			Desc:      "authentication configuration for multiple auth mechanisms",
		},
		{
			Name:     "meta",
			AttrType: TfMap,
			Optional: true,
			SubType:  TfString,
			Desc:     "metadata for the user",
		},
		{
			Name:      "tokens",
			AttrType:  TfJSON,
			Optional:  true,
			Sensitive: true,
			Desc:      "user tokens configuration",
		},
	}
}

const userConfigAttrDesc = "configuration for the user"

func dsUserConfigSchema() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		Attributes:          DSAttributes(userConfigAttrs()),
		Computed:            true,
		Description:         userConfigAttrDesc,
		MarkdownDescription: userConfigAttrDesc,
	}
}

func resUserConfigSchema() rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Attributes:          ResAttributes(userConfigAttrs()),
		Optional:            true,
		Description:         userConfigAttrDesc,
		MarkdownDescription: userConfigAttrDesc,
	}
}

func userAttrs() []BaseSchema {
	attrs := giveCommonAttributes()
	userAttrs := []BaseSchema{
		{
			Name:     "email",
			AttrType: TfString,
			Required: true,
			Desc:     "email address of the user",
		},
		{
			Name:     "active",
			AttrType: TfBoolean,
			Optional: true,
			Desc:     "whether the user is active",
		},
	}

	userAttrs = append(attrs, userAttrs...)

	return userAttrs
}

func UserDSchema() dschema.Schema {
	attrs := DSAttributes(userAttrs())
	attrs["config"] = dsUserConfigSchema()

	return dschema.Schema{
		Attributes:  attrs,
		Description: "user resource",
	}
}

func UserResourceSchema() rschema.Schema {
	attrs := ResAttributes(userAttrs())
	attrs["config"] = resUserConfigSchema()

	return rschema.Schema{
		Attributes:  attrs,
		Description: "user resource",
	}
}
