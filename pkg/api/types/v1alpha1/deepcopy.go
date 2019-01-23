package v1alpha1

import "k8s.io/apimachinery/pkg/runtime"

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *Populator) DeepCopyInto(out *Populator) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = PopulatorSpec{
		Type: in.Spec.Type,
	}
}

// DeepCopyObject returns a generically typed copy of an object
func (in *Populator) DeepCopyObject() runtime.Object {
	out := Populator{}
	in.DeepCopyInto(&out)

	return &out
}

// DeepCopyObject returns a generically typed copy of an object
func (in *PopulatorList) DeepCopyObject() runtime.Object {
	out := PopulatorList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]Populator, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}
