package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/yokecd/yoke/pkg/flight"

	v2 "github.com/yokecd/examples/demos/ingress-atc/backend/v2"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/utils/ptr"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var backend v2.Backend
	if err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&backend); err != nil && err != io.EOF {
		return fmt.Errorf("failed to unmarshal input as backend: %w", err)
	}

	selector := map[string]string{
		"app.kubernetes.io/name": backend.Name,
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: appsv1.SchemeGroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: backend.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(cmp.Or(backend.Spec.Replicas, 2)),
			Selector: &metav1.LabelSelector{
				MatchLabels: selector,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   backend.Name,
					Labels: selector,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "main",
							Image:   backend.Spec.Image,
							Command: backend.Spec.Command,
							Env: func() []corev1.EnvVar {
								var result []corev1.EnvVar
								for name, env := range backend.Spec.Env {
									if env.PlainText != "" {
										result = append(result, corev1.EnvVar{
											Name:  name,
											Value: env.PlainText,
										})
										continue
									}
									result = append(result, corev1.EnvVar{
										Name: name,
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												LocalObjectReference: corev1.LocalObjectReference{Name: env.Secret.Name},
												Key:                  env.Secret.Key,
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

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: corev1.SchemeGroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: backend.Name,
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
		},
	}

	ingress := func() *networkingv1.Ingress {
		if backend.Spec.PathPrefix == "" {
			return nil
		}

		return &networkingv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: networkingv1.SchemeGroupVersion.Identifier(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: backend.Name,
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										PathType: ptr.To(networkingv1.PathTypePrefix),
										Path:     backend.Spec.PathPrefix,
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: svc.Name,
												Port: networkingv1.ServiceBackendPort{
													Name: "http",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
	}()

	return json.NewEncoder(os.Stdout).Encode(flight.Resources{
		deployment,
		svc,
		ingress,
	})
}
