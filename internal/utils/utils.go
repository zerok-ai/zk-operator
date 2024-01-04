package utils

import (
	"reflect"
)

var LOG_TAG_UTILS = "utils"

func GetTypeName(i interface{}) string {
	// Use reflect.TypeOf().Name() to get the type name.
	return reflect.TypeOf(i).Name()
}
