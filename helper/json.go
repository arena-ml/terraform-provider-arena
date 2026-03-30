// Copyright (c) ArenaML Labs Pvt Ltd.

package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ConvertJSONStructToSimpleTF converts a source struct with pointer fields to a destination struct with Terraform types
// Matches fields by comparing json tags with tfsdk tags
// Supports string, int, int32, int64, bool field types
func ConvertJSONStructToSimpleTF[S any, U any](ctx context.Context, source S, dest *U) {
	sourceVal := reflect.ValueOf(source)
	sourceType := reflect.TypeOf(source)

	destVal := reflect.ValueOf(dest)
	destType := reflect.TypeOf(dest)
	if destType.Kind() != reflect.Ptr {
		tflog.Error(ctx, "expected dest to be a pointer")
		return
	}

	destVal = destVal.Elem()
	destType = destType.Elem()

	numFields := destType.NumField()
	// Iterate through destination struct fields
	for i := 0; i < numFields; i++ {
		destField := destType.Field(i)
		if destField.Type.Kind() == reflect.Ptr {
			// tflog.Warn(ctx, "DEST type is "+destField.Type.String()+" | "+destField.Name)
			continue
		} else {
			// tflog.Warn(ctx, "DEST type is "+destField.Type.String()+" | "+destField.Name)
		}
		destFieldVal := destVal.Field(i)

		// Skip unexported fields
		if !destFieldVal.CanSet() {
			tflog.Error(ctx, fmt.Sprintf("expected dest field to be setable %v", destFieldVal.Elem()))
			continue
		}

		// Get tfsdk tag from destination field
		tfsdkTag := getTfsdkTag(destField)
		if tfsdkTag == "" {
			// setTerraformNull(destFieldVal, destField.Type)
			// skip in case of tf tag is not set
			tflog.Warn(ctx, "tfsdk tag is required", map[string]interface{}{"dest_field": destField.Name})
			continue
		}

		// Find matching field in source struct by json tag
		sourceFieldIndex, err := findFieldIndexByTag(ctx, sourceType, jsonTagKey, tfsdkTag)
		if err != nil {
			// Set to null/zero value if source field not found
			// setTerraformNull(destFieldVal, destField.Type)
			tflog.Warn(ctx, err.Error())
			continue
		}

		sourceFieldVal := sourceVal.Field(sourceFieldIndex)

		// Convert and set the field value
		ok := convertAndSetTfField(ctx, sourceFieldVal, destFieldVal, destField.Type)
		if !ok {
			tflog.Warn(ctx, "Error converting field", map[string]interface{}{
				"dest_field": destField.Name,
				"dest_type":  destField.Type.String(),
				"sourceType": sourceFieldVal.Type(),
			})
		}
	}

	return
}

// getTfsdkTag extracts the tfsdk tag value from a struct field
func getTfsdkTag(field reflect.StructField) string {
	tag := field.Tag.Get("tfsdk")
	if tag == "" {
		return ""
	}

	// Handle tag options like "name,omitempty" - take only the first part
	if idx := strings.Index(tag, ","); idx != -1 {
		tag = tag[:idx]
	}

	return tag
}

// getJsonTag extracts the json tag value from a struct field
func getJsonTag(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return ""
	}

	// Handle tag options like "name,omitempty" - take only the first part
	if idx := strings.Index(tag, ","); idx != -1 {
		tag = tag[:idx]
	}

	return tag
}

type TagKey string

type getTagFn func(field reflect.StructField) string

const (
	tfsdkTagKey TagKey = "tfsdk"
	jsonTagKey  TagKey = "json"
)

func tagFnForKey(k TagKey) getTagFn {
	if k == tfsdkTagKey {
		return getTfsdkTag
	}

	if k == jsonTagKey {
		return getJsonTag
	}

	panic(fmt.Sprintf("unknown tag key: %s", k))
}

func findSourceFieldByTag(ctx context.Context, sourceVal reflect.Value, sourceType reflect.Type, tagName string, tagKey TagKey) (reflect.Value, bool) {
	if sourceType.Kind() == reflect.Ptr {
		tflog.Warn(ctx, "source is a pointer", map[string]interface{}{"source_type": sourceType.String(), "tfsdk_tag": tagName})
		return reflect.Value{}, false
	}

	getTag := tagFnForKey(tagKey)

	for i := 0; i < sourceType.NumField(); i++ {
		sourceField := sourceType.Field(i)
		jsonTag := getTag(sourceField)

		// Match the json tag with the tfsdk tag
		if jsonTag == tagName {
			return sourceVal.Field(i), true
		}
	}
	return reflect.Value{}, false
}

func findFieldIndexByTag(ctx context.Context, sourceType reflect.Type, tagKey TagKey, tagName string) (int, error) {
	if sourceType.Kind() == reflect.Ptr {
		tflog.Warn(ctx, "source is a pointer", map[string]interface{}{"source_type": sourceType.String(), "tfsdk_tag": tagName})
		return -1, fmt.Errorf("source is a pointer")
	}

	getTag := tagFnForKey(tagKey)

	for i := 0; i < sourceType.NumField(); i++ {
		sourceField := sourceType.Field(i)
		jsonTag := getTag(sourceField)

		// Match the json tag with the tfsdk tag
		if jsonTag == tagName {
			return i, nil
		}
	}

	return -1, fmt.Errorf("tag %s not found", tagName)
}

func findSourceFieldByName(sourceVal reflect.Value, sourceType reflect.Type, name string) (reflect.Value, bool) {
	if sourceType.Kind() == reflect.Ptr {
		return reflect.Value{}, false
	}

	for i := 0; i < sourceType.NumField(); i++ {
		sourceField := sourceType.Field(i)

		// Match the json tag with the tfsdk tag
		if sourceField.Name == name {
			return sourceVal.Field(i), true
		}
	}
	return reflect.Value{}, false
}

// convertAndSetTfField converts a source field value to the appropriate Terraform type
func convertAndSetTfField(ctx context.Context, sourceFieldVal reflect.Value, destFieldVal reflect.Value, destType reflect.Type) bool {
	// Handle pointer fields in source
	isNil := false
	if sourceFieldVal.Kind() == reflect.Ptr {
		if sourceFieldVal.IsNil() {
			// if source value is null then nothing need to be done as dest is already null
			isNil = true
			// return true
		}
		sourceFieldVal = sourceFieldVal.Elem()
	}

	switch destType {
	case tfStringType():
		if sourceFieldVal.Kind() == reflect.String {
			if isNil {
				destFieldVal.Set(reflect.ValueOf(types.StringNull()))
			} else {
				destFieldVal.Set(reflect.ValueOf(types.StringValue(sourceFieldVal.String())))
			}
			return true
		}
		tflog.Warn(ctx, fmt.Sprintf("type mismatch %s %s", sourceFieldVal.Kind(), destType.String()))
		return false

	case tfBoolType():
		if sourceFieldVal.Kind() == reflect.Bool {
			if isNil {
				destFieldVal.Set(reflect.ValueOf(types.BoolNull()))
			} else {
				destFieldVal.Set(reflect.ValueOf(types.BoolValue(sourceFieldVal.Bool())))
			}
			return true
		} else {
			tflog.Warn(ctx, fmt.Sprintf("type mismatch %s %s", sourceFieldVal.Kind(), destType.String()))
			return false
		}
	case tfInt64Type():
		switch sourceFieldVal.Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
			if isNil {
				destFieldVal.Set(reflect.ValueOf(types.Int64Null()))
			} else {
				destFieldVal.Set(reflect.ValueOf(types.Int64Value(sourceFieldVal.Int())))
			}
			return true
		default:
			tflog.Warn(ctx, fmt.Sprintf("type mismatch %s %s", sourceFieldVal.Kind(), destType.String()))
			return false
		}
	case tfInt32Type():
		switch sourceFieldVal.Kind() {
		case reflect.Int, reflect.Int32:
			if isNil {
				destFieldVal.Set(reflect.ValueOf(types.Int32Null()))
			} else {
				destFieldVal.Set(reflect.ValueOf(types.Int32Value(int32(sourceFieldVal.Int()))))
			}
			return true
		default:
			tflog.Warn(ctx, fmt.Sprintf("type mismatch %s %s", sourceFieldVal.Kind(), destType.String()))
			return false
		}
	case tfFloat32Type():
		switch sourceFieldVal.Kind() {
		case reflect.Float32:
			destFieldVal.Set(reflect.ValueOf(types.Float32Value(float32(sourceFieldVal.Float()))))
		default:
			tflog.Warn(ctx, fmt.Sprintf("type mismatch %s %s", sourceFieldVal.Kind(), destType.String()))
			return false
		}
	case tfFloat64Type():
		switch sourceFieldVal.Kind() {
		case reflect.Float32, reflect.Float64:
			destFieldVal.Set(reflect.ValueOf(types.Float64Value(sourceFieldVal.Float())))
			return true
		default:
			tflog.Warn(ctx, fmt.Sprintf("type mismatch %s %s", sourceFieldVal.Kind(), destType.String()))
			return false
		}
	case tfFloat64Type():
		if sourceFieldVal.Kind() == reflect.Float32 {
			destFieldVal.Set(reflect.ValueOf(types.Float32Value(float32(sourceFieldVal.Float()))))
			return true
		} else {
			tflog.Warn(ctx, fmt.Sprintf("type mismatch %s %s", sourceFieldVal.Kind(), destType.String()))
			return false
		}
	default:
		// not the type we can handle
		return false
	}

	return false
}

func tfStringType() reflect.Type {
	return reflect.TypeOf(basetypes.StringValue{})
}

func tfBoolType() reflect.Type {
	return reflect.TypeOf(basetypes.BoolValue{})
}

func tfInt64Type() reflect.Type {
	return reflect.TypeOf(basetypes.Int64Value{})
}

func tfInt32Type() reflect.Type {
	return reflect.TypeOf(basetypes.Int32Value{})
}

func tfFloat32Type() reflect.Type {
	return reflect.TypeOf(basetypes.Float32Value{})
}

func tfFloat64Type() reflect.Type {
	return reflect.TypeOf(basetypes.Float64Value{})
}

func JSONObjToStr(obj interface{}) (string, error) {
	jb, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(jb), nil
}

func JSONObjToNormalized(obj interface{}) (jsontypes.Normalized, error) {
	if obj == nil {
		return jsontypes.NewNormalizedNull(), nil
	}
	str, err := JSONObjToStr(obj)
	if err != nil {
		return jsontypes.Normalized{}, err
	}

	return jsontypes.NewNormalizedValue(str), nil
}

func MapConvertTfToJSON[S any, T any](ctx context.Context, s map[string]S, t map[string]T) error {
	for k, vTF := range s {
		var vGo T
		err := ConvertTfModelToApiJSON(ctx, vTF, &vGo)
		if err != nil {
			return err
		}

		t[k] = vGo
	}

	return nil
}

func MapConvertJsonToTf[S any, T any](ctx context.Context, s map[string]S, t map[string]T) {
	for k, vTF := range s {
		var vGo T
		ConvertJSONStructToSimpleTF(ctx, vTF, &vGo)
		t[k] = vGo
	}
}
