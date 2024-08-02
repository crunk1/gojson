package json

import (
	"fmt"
	"reflect"
)

// checkStructField checks:
// - optional and nullable tags are not used with omitempty tag
// - optional and nullable fields have enough indirection to represent optional and nullable values
func checkStructField(structType reflect.Type, f *field) error {
	requiredIndirectLevel := 0
	if f.optional {
		requiredIndirectLevel++
		if f.omitEmpty {
			return fmt.Errorf("json: field %q cannot have both omitempty and optional tags", f.name)
		}
	}
	if f.nullable {
		requiredIndirectLevel++
		if f.omitEmpty {
			return fmt.Errorf("json: field %q cannot have both omitempty and nullable tags", f.name)
		}
	}
	if requiredIndirectLevel == 0 {
		return nil // no required indirection for optional/nullable handling
	}

	// Find the field.
	fieldType := structType
	for _, i := range f.index {
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		fieldType = fieldType.Field(i).Type
	}

	// Check if the field has enough indirection.
	ft := fieldType
	for ft.Kind() == reflect.Ptr && requiredIndirectLevel > 0 {
		ft = ft.Elem()
		requiredIndirectLevel--
	}
	if requiredIndirectLevel > 0 {
		if f.optional && f.nullable {
			return fmt.Errorf("json: optional nullable field %q requires 2+ levels of indirection, type = %q", f.name, fieldType.String())
		}
		if f.optional {
			return fmt.Errorf("json: optional field %q requires 1+ levels of indirection, type = %q", f.name, fieldType.String())
		}
		if f.nullable {
			return fmt.Errorf("json: nullable field %q requires 1+ levels of indirection, type = %q", f.name, fieldType.String())
		}
	}
	return nil
}
