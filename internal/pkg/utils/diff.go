package utils

import (
	"reflect"

	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// CalculateDiff takes two structs (of the same type) and returns a map of changed fields.
func CalculateDiff(oldObj, newObj any) entity.DiffMap {
	diff := make(entity.DiffMap)
	if oldObj == nil || newObj == nil {
		return diff
	}

	vOld := reflect.ValueOf(oldObj)
	vNew := reflect.ValueOf(newObj)

	// Handle pointers
	if vOld.Kind() == reflect.Ptr {
		vOld = vOld.Elem()
	}
	if vNew.Kind() == reflect.Ptr {
		vNew = vNew.Elem()
	}

	// Ensure they are both structs
	if vOld.Kind() != reflect.Struct || vNew.Kind() != reflect.Struct {
		return diff
	}

	t := vOld.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		
		// Skip internal/unexported fields
		if f.PkgPath != "" {
			continue
		}

		fOld := vOld.Field(i)
		fNew := vNew.Field(i)

		// Get JSON tag or field name
		key := f.Tag.Get("json")
		if key == "" || key == "-" {
			key = f.Name
		}
		
		// Handle "field,omitempty"
		if commaIdx := findComma(key); commaIdx != -1 {
			key = key[:commaIdx]
		}

		if !reflect.DeepEqual(fOld.Interface(), fNew.Interface()) {
			diff[key] = entity.FieldDiff{
				Old: fOld.Interface(),
				New: fNew.Interface(),
			}
		}
	}

	return diff
}

func findComma(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			return i
		}
	}
	return -1
}
