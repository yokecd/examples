package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	v1 "github.com/yokecd/examples/demos/dynamic-mode/backend/v1"
	"github.com/yokecd/yoke/pkg/flight"
	"github.com/yokecd/yoke/pkg/flight/wasi/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
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

	externalSecret := &v1beta1.ExternalSecret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1beta1.SchemeGroupVersion.Identifier(),
			Kind:       "ExternalSecret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: backend.Name,
		},
		Spec: v1beta1.ExternalSecretSpec{
			SecretStoreRef: v1beta1.SecretStoreRef{
				Name: "vault-backend",
				Kind: "SecretStore",
			},
			Target: v1beta1.ExternalSecretTarget{
				Name:           secret.Name,
				CreationPolicy: v1beta1.CreatePolicyMerge,
				DeletionPolicy: v1beta1.DeletionPolicyRetain,
			},
			Data: func() []v1beta1.ExternalSecretData {
				var result []v1beta1.ExternalSecretData
				for _, value := range backend.Spec.Secrets {
					result = append(result, v1beta1.ExternalSecretData{
						SecretKey: value.Key,
						RemoteRef: v1beta1.ExternalSecretDataRemoteRef{
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
