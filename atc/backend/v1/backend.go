package v1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Backend is the type representing our CustomResource.
// It contains Type and Object meta as found in typical kubernetes objects and a spec.
// Do not provide a Status Object as that is automatically generated by the ATC.
type Backend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BackendSpec `json:"spec"`
}

// Our backend specification.
type BackendSpec struct {
	Image       string            `json:"image"`
	Replicas    int32             `json:"replicas"`
	Labels      map[string]string `json:"labels,omitempty"`
	NodePort    int               `json:"nodePort,omitempty"`
	ServicePort int               `json:"port,omitempty"`
}

// Marshalling helper to avoid needing to fill the Type meta which is already specific to this type and package.
func (backend Backend) MarshalJSON() ([]byte, error) {
	backend.Kind = "Backend"
	backend.APIVersion = "examples.com/v1"

	type BackendAlt Backend
	return json.Marshal(BackendAlt(backend))
}
