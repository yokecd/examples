package v1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/yokecd/yoke/pkg/flight"
)

var SchemeGroupVersion = schema.GroupVersion{
	Group:   "examples.com",
	Version: "v1",
}

type Backend struct {
	metav1.TypeMeta
	metav1.ObjectMeta `json:"metadata"`
	Spec              BackendSpec   `json:"spec,omitzero"`
	Status            flight.Status `json:"status,omitzero"`
}

type BackendSpec struct {
	Image      string            `json:"image"`
	Command    []string          `json:"command,omitempty"`
	PathPrefix string            `json:"pathPrefix,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
}

func (backend Backend) MarshalJSON() ([]byte, error) {
	backend.APIVersion = SchemeGroupVersion.Identifier()
	backend.Kind = "Backend"

	type BackendAlt Backend
	return json.Marshal(BackendAlt(backend))
}
