package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type GitPopulator struct {
	Repo       string `json:"repo"`
	Branch     string `json:"branch"`
	Mountpoint string `json:"mountpoint"`
}

type PopulatorSpec struct {
	Type string       `json:"type"`
	Git  GitPopulator `json:"git"`
}

type Populator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PopulatorSpec `json:"spec"`
}

type PopulatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Populator `json:"items"`
}
