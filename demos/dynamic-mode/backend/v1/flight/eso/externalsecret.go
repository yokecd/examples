package eso

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ExternalSecret represents a vault specfic version of "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1":ExternalSecret
//
// The reason we are not importing that library directly, is that there is some bug with the Go Toolchain
// Where when this package is included it fails to compile wasmexport directives properly.
// This bug is included here: https://github.com/issues/created?issue=golang%7Cgo%7C73246
type ExternalSecret struct {
	metav1.TypeMeta
	metav1.ObjectMeta `json:"metadata"`
	Spec              ExternalSecretSpec `json:"spec"`
}

type ExternalSecretSpec struct {
	RefreshInterval *metav1.Duration       `json:"refreshInterval,omitzero"`
	SecretStoreRef  ExternalSecretStoreRef `json:"secretStoreRef"`
	Target          ExternalSecretTarget   `json:"target"`
	Data            []ExternalSecretData   `json:"data"`
}

type ExternalSecretStoreRef struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type ExternalSecretTarget struct {
	Name           string `json:"name"`
	CreationPolicy string `json:"creationPolicy"`
}

type ExternalSecretData struct {
	SecretKey string    `json:"secretKey"`
	RemoteRef RemoteRef `json:"remoteRef"`
}

type RemoteRef struct {
	Key      string `json:"key"`
	Property string `json:"property"`
}
