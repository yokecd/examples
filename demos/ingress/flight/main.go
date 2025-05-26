package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/yokecd/yoke/pkg/flight"

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

type Config struct {
	Name       string            `json:"-"`
	Image      string            `json:"image"`
	Command    []string          `json:"command"`
	Replicas   int32             `json:"replicas"`
	PathPrefix string            `json:"pathPrefix"`
	Env        map[string]string `json:"env"`
}

func run() error {
	cfg := Config{
		Name:  flight.Release(),
		Image: "ealen/echo-server:latest",
	}

	if err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&cfg); err != nil && err != io.EOF {
		return fmt.Errorf("failed to unmarshal input into expected config: %w", err)
	}

	selector := map[string]string{
		"app.kubernetes.io/name": cfg.Name,
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: appsv1.SchemeGroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: cfg.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(cmp.Or(cfg.Replicas, 2)),
			Selector: &metav1.LabelSelector{
				MatchLabels: selector,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   cfg.Name,
					Labels: selector,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "main",
							Image:   cfg.Image,
							Command: cfg.Command,
							Env: func() []corev1.EnvVar {
								var result []corev1.EnvVar
								for key, value := range cfg.Env {
									result = append(result, corev1.EnvVar{
										Name:  key,
										Value: value,
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
			Name: cfg.Name,
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
		if cfg.PathPrefix == "" {
			return nil
		}

		return &networkingv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: networkingv1.SchemeGroupVersion.Identifier(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: cfg.Name,
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										PathType: ptr.To(networkingv1.PathTypePrefix),
										Path:     cfg.PathPrefix,
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
