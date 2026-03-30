// Copyright (c) ArenaML Labs Pvt Ltd.

package helper

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// FromGoStrSliceToTfList convert terraform string list to go string slice
func FromGoStrSliceToTfList(ctx context.Context, list []string) (types.List, error) {
	tfList, diag := types.ListValueFrom(ctx, types.StringType, list)
	if diag.HasError() {
		return tfList, fmt.Errorf("%s", diag.Errors()[0].Detail())
	}
	return tfList, nil
}

// FromGoStrSliceToTfSet convert terraform string list to go string slice
func FromGoStrSliceToTfSet(ctx context.Context, list []string) (types.Set, error) {
	if list == nil {
		return types.SetNull(types.StringType), nil
	}
	tfList, diag := types.SetValueFrom(ctx, types.StringType, list)
	if diag.HasError() {
		return tfList, fmt.Errorf("%s", diag.Errors()[0].Detail())
	}
	return tfList, nil
}

// ConvertTfModelToApiJSON convert the oapi-client struct to tf schema models
// S is the struct type with tfsdk tags, and T is a struct type with json tags
func ConvertTfModelToApiJSON[S any, T any](ctx context.Context, source S, target T) (err error) {
	// sourceVal := reflect.ValueOf(source)
	sourceVal := reflect.ValueOf(source)
	sourceVal = reflect.Indirect(sourceVal)
	sourceType := sourceVal.Type()

	targetVal := reflect.ValueOf(target)
	targetType := reflect.TypeOf(target)

	// Ensure target is a pointer so we can set values
	if targetVal.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	targetVal = targetVal.Elem()
	targetType = targetType.Elem()

	numFields := reflect.Indirect(sourceVal).NumField()

	for i := 0; i < numFields; i++ {
		sourceFieldVal := sourceVal.Field(i)

		if !ValidTFPrimitive(sourceFieldVal.Type()) {
			log.Println("invalid primitive type", sourceFieldVal)
			continue
		}
		log.Println("converting field", i, sourceFieldVal.Kind(), sourceFieldVal.Type())

		sourceField := sourceType.Field(i)

		tfKey := getTfsdkTag(sourceField)
		if tfKey == "" {
			log.Println("invalid tfsdk tag", sourceField.Name)
			continue
		}
		targetFieldIndex, err := findFieldIndexByTag(ctx, targetType, jsonTagKey, tfKey)
		if err != nil {
			tflog.Warn(ctx, fmt.Sprintf("unable to find field '%s' in dest ", tfKey), map[string]interface{}{"error": err})
			log.Println("unable to find field " + tfKey)
			continue
		}

		targetFieldVal := targetVal.Field(targetFieldIndex)
		if !targetFieldVal.CanSet() {
			log.Println("unable to set field ", tfKey)
			continue
		}
		targetFieldType := targetType.Field(targetFieldIndex)

		err = setGoValueFromTFPrimitive(sourceFieldVal, sourceField.Type, targetFieldVal, targetFieldType.Type)
		if err != nil {
			log.Println("unable to convert field ", tfKey)
			return err
		}
	}

	return nil
}

func ValidTFPrimitive(p reflect.Type) bool {
	if p == tfStringType() || p == tfBoolType() || p == tfInt32Type() || p == tfInt64Type() ||
		p == tfBoolType() || p == tfFloat32Type() || p == tfFloat64Type() {
		return true
	}

	return false
}

// setGoValueFromTFPrimitive check if tfVal is compatible with go type and accordingly sets the destVal.
func setGoValueFromTFPrimitive(tfVal reflect.Value, tfType reflect.Type, destVal reflect.Value, destType reflect.Type) error {
	// Handle null/unknown values
	switch tfType {
	case tfStringType():
		tfString, ok := tfVal.Interface().(types.String)
		if !ok {
			return fmt.Errorf("expected types.String, got %T", tfVal.Interface())
		}
		if tfString.IsNull() || tfString.IsUnknown() {
			return nil // Skip null/unknown values
		}

		stringVal := tfString.ValueString()
		if destType.Kind() == reflect.Ptr {
			// Handle pointer types
			if destType.Elem().Kind() == reflect.String {
				destVal.Set(reflect.ValueOf(&stringVal))
			} else {
				return fmt.Errorf("unsupported pointer type for string: %v", destType.Elem().Kind())
			}
		} else if destType.Kind() == reflect.String {
			destVal.SetString(stringVal)
		} else {
			return fmt.Errorf("cannot convert types.String to %v", destType.Kind())
		}

	case tfBoolType():
		tfBool, ok := tfVal.Interface().(types.Bool)
		if !ok {
			return fmt.Errorf("expected types.Bool, got %T", tfVal.Interface())
		}
		if tfBool.IsNull() || tfBool.IsUnknown() {
			return nil
		}

		boolVal := tfBool.ValueBool()
		if destType.Kind() == reflect.Ptr {
			if destType.Elem().Kind() == reflect.Bool {
				destVal.Set(reflect.ValueOf(&boolVal))
			} else {
				return fmt.Errorf("unsupported pointer type for bool: %v", destType.Elem().Kind())
			}
		} else if destType.Kind() == reflect.Bool {
			destVal.SetBool(boolVal)
		} else {
			return fmt.Errorf("cannot convert types.Bool to %v", destType.Kind())
		}

	case tfInt32Type():
		tfInt32, ok := tfVal.Interface().(types.Int32)
		if !ok {
			return fmt.Errorf("expected types.Int32, got %T", tfVal.Interface())
		}
		if tfInt32.IsNull() || tfInt32.IsUnknown() {
			return nil
		}

		int32Val := tfInt32.ValueInt32()
		if destType.Kind() == reflect.Ptr {
			if destType.Elem().Kind() == reflect.Int32 {
				destVal.Set(reflect.ValueOf(&int32Val))
			} else {
				if destType.Elem().Kind() == reflect.Int {
					iVal := int(int32Val)
					destVal.Set(reflect.ValueOf(&iVal))
				} else {
					return fmt.Errorf("unsupported pointer type for int32: %v", destType.Elem().Kind())
				}
			}
		} else if destType.Kind() == reflect.Int32 {
			destVal.SetInt(int64(int32Val))
		} else {
			return fmt.Errorf("cannot convert types.Int32 to %v", destType.Kind())
		}

	case tfInt64Type():
		tfInt64, ok := tfVal.Interface().(types.Int64)
		if !ok {
			return fmt.Errorf("expected types.Int64, got %T", tfVal.Interface())
		}
		if tfInt64.IsNull() || tfInt64.IsUnknown() {
			return nil
		}

		int64Val := tfInt64.ValueInt64()
		if destType.Kind() == reflect.Ptr {
			if destType.Elem().Kind() == reflect.Int64 {
				destVal.Set(reflect.ValueOf(&int64Val))
			} else {
				return fmt.Errorf("unsupported pointer type for int64: %v", destType.Elem().Kind())
			}
		} else if destType.Kind() == reflect.Int64 || destType.Kind() == reflect.Int {
			destVal.SetInt(int64Val)
		} else {
			return fmt.Errorf("cannot convert types.Int64 to %v", destType.Kind())
		}

	case tfFloat32Type():
		tfFloat32, ok := tfVal.Interface().(types.Float32)
		if !ok {
			return fmt.Errorf("expected types.Float32, got %T", tfVal.Interface())
		}
		if tfFloat32.IsNull() || tfFloat32.IsUnknown() {
			return nil
		}

		float32Val := tfFloat32.ValueFloat32()
		if destType.Kind() == reflect.Ptr {
			if destType.Elem().Kind() == reflect.Float32 {
				destVal.Set(reflect.ValueOf(&float32Val))
			} else {
				return fmt.Errorf("unsupported pointer type for float32: %v", destType.Elem().Kind())
			}
		} else if destType.Kind() == reflect.Float32 {
			destVal.SetFloat(float64(float32Val))
		} else {
			return fmt.Errorf("cannot convert types.Float32 to %v", destType.Kind())
		}

	case tfFloat64Type():
		tfFloat64, ok := tfVal.Interface().(types.Float64)
		if !ok {
			return fmt.Errorf("expected types.Float64, got %T", tfVal.Interface())
		}
		if tfFloat64.IsNull() || tfFloat64.IsUnknown() {
			return nil
		}

		float64Val := tfFloat64.ValueFloat64()
		if destType.Kind() == reflect.Ptr {
			if destType.Elem().Kind() == reflect.Float64 {
				destVal.Set(reflect.ValueOf(&float64Val))
			} else {
				return fmt.Errorf("unsupported pointer type for float64: %v", destType.Elem().Kind())
			}
		} else if destType.Kind() == reflect.Float64 || destType.Kind() == reflect.Float32 {
			destVal.SetFloat(float64Val)
		} else {
			return fmt.Errorf("cannot convert types.Float64 to %v", destType.Kind())
		}

	default:
		return fmt.Errorf("unsupported terraform type: %v", tfType)
	}

	return nil
}

// ToTfStringMap convert a go map[string]string to tf map
func ToTfStringMap(ctx context.Context, gMap *map[string]string) (types.Map, error) {
	if gMap == nil {
		return types.MapNull(types.StringType), nil
	}

	tfMap, diag := types.MapValueFrom(ctx, types.StringType, *gMap)
	if diag.HasError() {
		return types.MapNull(types.StringType), fmt.Errorf("%s : %s", diag[0].Detail(), diag[0].Summary())
	}

	return tfMap, nil
}

func TfMapStrToGoMap(ctx context.Context, tm types.Map) (map[string]string, bool) {
	if tm.IsNull() || tm.IsUnknown() {
		return nil, true
	}
	var ra map[string]string
	diags := tm.ElementsAs(ctx, &ra, false)
	if diags.HasError() {
		return nil, false
	}
	return ra, true
}

func TfListStrToGoSlice(ctx context.Context, tm types.List) ([]string, bool) {
	if tm.IsNull() || tm.IsUnknown() {
		return nil, true
	}
	var l []string
	diags := tm.ElementsAs(ctx, &l, false)
	if diags.HasError() {
		return nil, false
	}
	return l, true
}

func TfSetStrToGoSlice(ctx context.Context, tm types.Set) ([]string, bool) {
	if tm.IsNull() || tm.IsUnknown() {
		return []string{}, true
	}
	var l []string
	diags := tm.ElementsAs(ctx, &l, false)
	if diags.HasError() {
		return nil, false
	}
	return l, true
}

func TfJSONToGoMapInterface(ctx context.Context, tm jsontypes.Normalized) (map[string]interface{}, error) {
	var spec map[string]interface{}
	diag := tm.Unmarshal(&spec)
	if diag.HasError() {
		return nil, fmt.Errorf("\n%s", diag.Errors()[0].Detail())
	}

	return spec, nil
}

func GoMapFloat64FromTfMap(ctx context.Context, tm types.Map) (map[string]float64, bool) {
	if tm.IsNull() || tm.IsUnknown() {
		return nil, true
	}
	var ra map[string]float64
	diags := tm.ElementsAs(ctx, &ra, false)
	if diags.HasError() {
		return nil, false
	}
	return ra, true
}

func TfMapFromGoMapFloat64(ctx context.Context, gMap *map[string]float64) (types.Map, error) {
	if gMap == nil {
		return types.MapNull(types.Float64Type), nil
	}

	tfMap, diag := types.MapValueFrom(ctx, types.Float64Type, *gMap)
	if diag.HasError() {
		return types.MapNull(types.Float64Type), fmt.Errorf("%s : %s", diag[0].Detail(), diag[0].Summary())
	}

	return tfMap, nil
}

func GoMapFloat32FromTfMap(ctx context.Context, tm types.Map) (map[string]float32, bool) {
	if tm.IsNull() || tm.IsUnknown() {
		return nil, true
	}
	var ra map[string]float32
	diags := tm.ElementsAs(ctx, &ra, false)
	if diags.HasError() {
		return nil, false
	}
	return ra, true
}

func TfMapFromGoMapFloat32(ctx context.Context, gMap *map[string]float32) (types.Map, error) {
	if gMap == nil {
		return types.MapNull(types.Float32Type), nil
	}

	tfMap, diag := types.MapValueFrom(ctx, types.Float32Type, *gMap)
	if diag.HasError() {
		return types.MapNull(types.Float32Type), fmt.Errorf("%s : %s", diag[0].Detail(), diag[0].Summary())
	}

	return tfMap, nil
}
