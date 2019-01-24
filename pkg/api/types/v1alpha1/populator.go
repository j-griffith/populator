package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// GitPopulator provides a struct with the specific git information that might be desired
type GitPopulator struct {
	Repo   string `json:"repo"` // Full URL of the repo (https or git protocol)
	Branch string `json:"branch"`
	Tag    string `json:"tag,omitempty"`
}

// PopulatorSpec provides a struct that details the type of external data source we're working with, as well as where to mount
// the data we're populating (ie root directory).  We also provide a mechanism to override the built in container images with
// your own custom images.  Be warned, it's up to you to make sure you have proper enetry points etc here
type PopulatorSpec struct {
	//ImageOverride string       `json:"image_override"`
	SecretRef  string       `json:"secret_ref"`
	Type       string       `json:"type"`
	Mountpoint string       `json:"mountpoint"`
	Git        GitPopulator `json:"git"`
}

// Populator represents our CRD Object.  A populator is a DataSource used to pre-populate PVCs upon creation
type Populator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PopulatorSpec `json:"spec"`
}

// PopulatorList provides a type of multiple Populators
type PopulatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Populator `json:"items"`
}
