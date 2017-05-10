package validator

import "reflect"

// Return true when all fields is zero.
// sIface must be pointer to struct.
func IsStructZero(sIface interface{}) bool {
	v := reflect.ValueOf(sIface)
	if v.Kind() != reflect.Ptr {
		panic("sIface must be pointer to struct")
	}
	if v.IsNil() || !v.IsValid() {
		panic("sIface must be pointer to struct")
	}

	vv := v.Elem()
	for i := 0; i < vv.NumField(); i++ {
		field := vv.Field(i)
		if field.Interface() != reflect.Zero(field.Type()).Interface() {
			return false
		}
	}

	return true
}
