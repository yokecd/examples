package v1

import (
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/yokecd/yoke/pkg/openapi"
)

const (
	APIVersion  = "examples.com/v1"
	KindBackend = "Backend"
)

// Backend is the type representing our CustomResource.
// It contains Type and Object meta as found in typical kubernetes objects and a spec.
// Do not provide a Status Object as that is automatically generated by the ATC.
type Backend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitzero"`
	Spec              BackendSpec `json:"spec"`
}

type Secrets map[string]struct {
	Path string `json:"path"`
	Key  string `json:"key"`
}

// Our Backend Specification
type BackendSpec struct {
	Image                  string           `json:"image"`
	Replicas               int32            `json:"replicas"`
	Secrets                Secrets          `json:"secrets,omitempty"`
	SecretRefreshInternval openapi.Duration `json:"refreshInterval,omitzero"`
}

// Custom Marshalling Logic so that users do not need to explicity fill out the Kind and ApiVersion.
func (backend Backend) MarshalJSON() ([]byte, error) {
	backend.Kind = KindBackend
	backend.APIVersion = APIVersion

	type BackendAlt Backend
	return json.Marshal(BackendAlt(backend))
}

// Custom Unmarshalling to raise an error if the ApiVersion or Kind does not match.
func (backend *Backend) UnmarshalJSON(data []byte) error {
	type BackendAlt Backend
	if err := json.Unmarshal(data, (*BackendAlt)(backend)); err != nil {
		return err
	}
	if backend.APIVersion != APIVersion {
		return fmt.Errorf("unexpected api version: expected %s but got %s", APIVersion, backend.APIVersion)
	}
	if backend.Kind != KindBackend {
		return fmt.Errorf("unexpected kind: expected %s but got %s", KindBackend, backend.Kind)
	}
	return nil
}
