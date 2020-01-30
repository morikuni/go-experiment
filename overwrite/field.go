package overwrite

import (
	"errors"
	"reflect"
	"unsafe"
)

var (
	ErrNotStruct    = errors.New("not struct")
	ErrTypeMismatch = errors.New("type mismatch")
	ErrNoSuchField  = errors.New("no such field")
)

// Field override the value of `target.field` to val.
// PLEASE DO NOT USE THIS FUNCTION IN YOUR PRODUCTION CODE.
func Field(target interface{}, field string, val interface{}) error {
	targetVal := reflect.Indirect(reflect.ValueOf(target))

	if targetVal.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	dstVal := targetVal.FieldByName(field)
	if !dstVal.IsValid() {
		return ErrNoSuchField
	}

	srcVal := reflect.ValueOf(val)
	srcType := srcVal.Type()
	dstType := dstVal.Type()
	if srcType != dstType {
		if dstVal.Kind() == reflect.Interface && srcType.Implements(dstType) {
			srcVal = srcVal.Convert(dstType)
		} else {
			return ErrTypeMismatch
		}
	}

	dstAddr := unsafe.Pointer(dstVal.UnsafeAddr())

	setableField := reflect.NewAt(dstVal.Type(), dstAddr).Elem()

	setableField.Set(srcVal)

	return nil
}
