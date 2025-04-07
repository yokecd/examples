package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	esmeta "github.com/external-secrets/external-secrets/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/utils/ptr"

	"github.com/yokecd/yoke/pkg/flight"

	"github.com/yokecd/examples/internal/charts/eso"
	"github.com/yokecd/examples/internal/charts/vault"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Config struct {
	Vault vault.Values `json:"vault"`
	ESO   eso.Values   `json:"eso"`
}

func run() error {
	cfg := Config{
		Vault: vault.Values{
			Server: &vault.ValuesServer{
				Dev: &vault.ValuesServerDev{
					Enabled: ptr.To(true),
				},
				ReadinessProbe: &vault.ValuesServerReadinessProbe{
					Enabled: ptr.To(false),
				},
				LivenessProbe: &vault.ValuesServerLivenessProbe{
					Enabled: ptr.To(false),
				},
			},
			Global: &vault.ValuesGlobal{
				Enabled:    ptr.To(true),
				TlsDisable: ptr.To(false),
			},
		},
		ESO: eso.Values{},
	}

	if err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&cfg); err != nil && err != io.EOF {
		return fmt.Errorf("failed to parse input: %v", err)
	}

	vaultResources, err := vault.RenderChart(flight.Release()+"-vault", flight.Namespace(), &cfg.Vault)
	if err != nil {
		return fmt.Errorf("failed to render vault chart: %v", err)
	}

	esoResources, err := eso.RenderChart(flight.Release()+"-eso", flight.Namespace(), &cfg.ESO)
	if err != nil {
		return fmt.Errorf("failed to render eso chart: %v", err)
	}

	var resources flight.Resources
	for _, resource := range append(vaultResources, esoResources...) {
		resources = append(resources, resource)
	}

	vaultTokenSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "vault-token",
		},
		StringData: map[string]string{"token": "root"},
		Type:       corev1.SecretTypeOpaque,
	}

	vaultBackend := &v1beta1.SecretStore{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1beta1.SchemeGroupVersion.Identifier(),
			Kind:       "SecretStore",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "vault-backend",
		},
		Spec: v1beta1.SecretStoreSpec{
			Provider: &v1beta1.SecretStoreProvider{
				Vault: &v1beta1.VaultProvider{
					Server:  fmt.Sprintf("http://%s-vault:8200", flight.Release()),
					Path:    ptr.To("secret"),
					Version: v1beta1.VaultKVStoreV2,
					Auth: &v1beta1.VaultAuth{
						TokenSecretRef: &esmeta.SecretKeySelector{
							Name: "vault-token",
							Key:  "token",
						},
					},
				},
			},
		},
	}

	resources = append(resources, vaultTokenSecret, vaultBackend)

	var crds, other flight.Resources
	for _, resource := range resources {
		if resource.GroupVersionKind().Kind == "CustomResourceDefinition" {
			crds = append(crds, resource)
		} else {
			other = append(other, resource)
		}
	}

	return json.NewEncoder(os.Stdout).Encode(flight.Stages{crds, other})
}
