package utils

import (
	"reflect"
)

var LOG_TAG_UTILS = "utils"

func RespCodeIsOk(status int) bool {
	if status > 199 && status < 300 {
		return true
	}
	return false

}

func GetTypeName(i interface{}) string {
	// Use reflect.TypeOf().Name() to get the type name.
	return reflect.TypeOf(i).Name()
}
