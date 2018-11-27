package overwrite

import (
	"errors"
	"reflect"
	"unsafe"
)

var (
	ErrTypeMismatch = errors.New("type mismatch")
	ErrNoSuchField  = errors.New("no such field")
)

// Field override the value of `target.field` to val.
func Field(target interface{}, field string, val interface{}) error {
	targetVal := reflect.Indirect(reflect.ValueOf(target))

	srcVal := reflect.ValueOf(val)
	dstVal := targetVal.FieldByName(field)

	if !dstVal.IsValid() {
		return ErrNoSuchField
	}

	if srcVal.Type() != dstVal.Type() {
		return ErrTypeMismatch
	}

	dstAddr := unsafe.Pointer(dstVal.UnsafeAddr())

	setableField := reflect.NewAt(dstVal.Type(), dstAddr).Elem()

	setableField.Set(srcVal)

	return nil
}
