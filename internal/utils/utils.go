package utils

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
)

const (
	Group               = "operator.zerok.ai"
	Version             = "v1alpha1"
	ZeroKProbeResource  = "zerokprobes"
	ZeroKProbeKind      = "ZerokProbe"
	ServiceListRedisKey = "serviceList"
)

func GetTypeName(i interface{}) string {
	return reflect.TypeOf(i).Name()
}

func SchemaGroupVersionKindForResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: ZeroKProbeResource,
	}
}
