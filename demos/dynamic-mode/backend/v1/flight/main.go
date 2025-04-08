package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/yokecd/yoke/pkg/flight"
	"github.com/yokecd/yoke/pkg/flight/wasi/k8s"

	v1 "github.com/yokecd/examples/demos/dynamic-mode/backend/v1"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var backend v1.Backend
	if err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&backend); err != nil && err != io.EOF {
		return fmt.Errorf("failed to decore backend: %v", err)
	}

	selector := map[string]string{"app.kubernetes.io/name": backend.Name}

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: backend.Name,
		},
		Data: func() map[string][]byte {
			value, err := k8s.Lookup[corev1.Secret](k8s.ResourceIdentifier{
				Name:       backend.Name,
				Namespace:  backend.Namespace,
				Kind:       "Secret",
				ApiVersion: "v1",
			})
			if err != nil || value == nil {
				return map[string][]byte{}
			}
			return value.Data
		}(),
	}

	externalSecret := &ExternalSecret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "external-secrets.io/v1beta1",
			Kind:       "ExternalSecret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: backend.Name,
		},
		Spec: Spec{
			RefreshInterval: func() *metav1.Duration {
				if backend.Spec.SecretRefreshInternval > 0 {
					return &metav1.Duration{Duration: time.Duration(backend.Spec.SecretRefreshInternval)}
				}
				return &metav1.Duration{Duration: 5 * time.Second}
			}(),
			SecretStoreRef: SecretStoreRef{
				Name: "vault-backend",
				Kind: "SecretStore",
			},
			Target: Target{
				Name:           secret.Name,
				CreationPolicy: "Merge",
			},
			Data: func() []Data {
				var result []Data
				for _, value := range backend.Spec.Secrets {
					result = append(result, Data{
						SecretKey: value.Key,
						RemoteRef: RemoteRef{
							Key:      value.Path,
							Property: value.Key,
						},
					})
				}
				return result
			}(),
		},
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.Identifier(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: backend.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &backend.Spec.Replicas,
			Selector: &metav1.LabelSelector{MatchLabels: selector},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selector,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  backend.Name,
							Image: backend.Spec.Image,
							Env: func() []corev1.EnvVar {
								var result []corev1.EnvVar
								for name, value := range backend.Spec.Secrets {
									result = append(result, corev1.EnvVar{
										Name: name,
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
												Key:                  value.Key,
											},
										},
									})
								}
								return result
							}(),
						},
					},
				},
			},
		},
	}

	return json.NewEncoder(os.Stdout).Encode(flight.Resources{deployment, secret, externalSecret})
}
