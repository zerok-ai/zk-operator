//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"github.com/zerok-ai/zk-utils-go/scenario/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Workloads) DeepCopyInto(out *Workloads) {
	{
		in := &in
		*out = make(Workloads, len(*in))
		//for key, val := range *in {
		//	(*out)[key] = *val.DeepCopy()
		//}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Workloads.
func (in Workloads) DeepCopy() Workloads {
	if in == nil {
		return nil
	}
	out := new(Workloads)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZerokProbe) DeepCopyInto(out *ZerokProbe) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZerokProbe.
func (in *ZerokProbe) DeepCopy() *ZerokProbe {
	if in == nil {
		return nil
	}
	out := new(ZerokProbe)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ZerokProbe) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZerokProbeSpec) DeepCopyInto(out *ZerokProbeSpec) {
	*out = *in
	if in.Workloads != nil {
		in, out := &in.Workloads, &out.Workloads
		*out = make(Workloads, len(*in))
		//for key, val := range *in {
		//	(*out)[key] = *val.DeepCopy()
		//}
	}
	//in.Filter.DeepCopyInto(&out.Filter)
	if in.GroupBy != nil {
		in, out := &in.GroupBy, &out.GroupBy
		*out = make([]model.GroupBy, len(*in))
		copy(*out, *in)
	}
	if in.RateLimit != nil {
		in, out := &in.RateLimit, &out.RateLimit
		*out = make([]model.RateLimit, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZerokProbeSpec.
func (in *ZerokProbeSpec) DeepCopy() *ZerokProbeSpec {
	if in == nil {
		return nil
	}
	out := new(ZerokProbeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZerokProbeStatus) DeepCopyInto(out *ZerokProbeStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZerokProbeStatus.
func (in *ZerokProbeStatus) DeepCopy() *ZerokProbeStatus {
	if in == nil {
		return nil
	}
	out := new(ZerokProbeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Zerokinstrumentation) DeepCopyInto(out *Zerokinstrumentation) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Zerokinstrumentation.
func (in *Zerokinstrumentation) DeepCopy() *Zerokinstrumentation {
	if in == nil {
		return nil
	}
	out := new(Zerokinstrumentation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Zerokinstrumentation) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZerokinstrumentationList) DeepCopyInto(out *ZerokinstrumentationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Zerokinstrumentation, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZerokinstrumentationList.
func (in *ZerokinstrumentationList) DeepCopy() *ZerokinstrumentationList {
	if in == nil {
		return nil
	}
	out := new(ZerokinstrumentationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ZerokinstrumentationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZerokinstrumentationSpec) DeepCopyInto(out *ZerokinstrumentationSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZerokinstrumentationSpec.
func (in *ZerokinstrumentationSpec) DeepCopy() *ZerokinstrumentationSpec {
	if in == nil {
		return nil
	}
	out := new(ZerokinstrumentationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZerokinstrumentationStatus) DeepCopyInto(out *ZerokinstrumentationStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZerokinstrumentationStatus.
func (in *ZerokinstrumentationStatus) DeepCopy() *ZerokinstrumentationStatus {
	if in == nil {
		return nil
	}
	out := new(ZerokinstrumentationStatus)
	in.DeepCopyInto(out)
	return out
}
