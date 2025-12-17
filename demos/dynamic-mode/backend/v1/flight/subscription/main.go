package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"time"

	esov1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1"

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

	externalSecret := &esov1.ExternalSecret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: esov1.SchemeGroupVersion.Identifier(),
			Kind:       "ExternalSecret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: backend.Name,
		},
		Spec: esov1.ExternalSecretSpec{
			RefreshInterval: func() *metav1.Duration {
				if backend.Spec.SecretRefreshInternval.Duration > 0 {
					return &backend.Spec.SecretRefreshInternval
				}
				return &metav1.Duration{Duration: 5 * time.Second}
			}(),
			SecretStoreRef: esov1.SecretStoreRef{
				Name: "vault-backend",
				Kind: "SecretStore",
			},
			Target: esov1.ExternalSecretTarget{
				Name:           backend.Name,
				CreationPolicy: esov1.CreatePolicyMerge,
			},
			Data: func() []esov1.ExternalSecretData {
				var result []esov1.ExternalSecretData
				for _, value := range backend.Spec.Secrets {
					result = append(result, esov1.ExternalSecretData{
						SecretKey: value.Key,
						RemoteRef: esov1.ExternalSecretDataRemoteRef{
							Key:      value.Path,
							Property: value.Key,
						},
					})
				}
				return result
			}(),
		},
	}

	secret, err := k8s.Lookup[corev1.Secret](k8s.ResourceIdentifier{
		ApiVersion: "v1",
		Kind:       "Secret",
		Name:       externalSecret.Spec.Target.Name,
		Namespace:  externalSecret.Namespace,
	})
	if err != nil && !k8s.IsErrNotFound(err) {
		return fmt.Errorf("failed to lookup secret: %w", err)
	}

	secretHash := func() string {
		if secret == nil {
			return ""
		}
		hash := sha1.New()
		for _, key := range slices.Sorted(maps.Keys(secret.Data)) {
			hash.Write(secret.Data[key])
		}
		return hex.EncodeToString(hash.Sum(nil))
	}()

	labels := map[string]string{"secret-hash": secretHash}

	maps.Copy(labels, selector)

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
					Labels: labels,
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

	return json.NewEncoder(os.Stdout).Encode(flight.Resources{deployment, externalSecret})
}
