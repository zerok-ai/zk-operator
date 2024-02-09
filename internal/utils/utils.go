package utils

import (
	"reflect"
)

func GetTypeName(i interface{}) string {
	return reflect.TypeOf(i).Name()
}
