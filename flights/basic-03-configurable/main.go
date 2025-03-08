package main

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"

	"github.com/yokecd/yoke/pkg/flight"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

// Keeping the helm nomenclature for familiarity, but you don't have to!
type Values struct {
	Replicas      int32             `json:"replicas"`
	ContainerPort int32             `json:"containerPort"`
	ServicePort   int32             `json:"servicePort"`
	Labels        map[string]string `json:"labels"`
}

var (
	release   = flight.Release()   // the first argument passed to yoke takeoff;       ie: yoke takeoff RELEASE foo
	namespace = flight.Namespace() // the value of the flag namespace during takeoff;  ie: yoke takeoff -namespace NAMESPACE ...
	selector  = map[string]string{"app.kubernetes.io/name": release}
)

func run() error {
	// Create values with defaults.
	values := Values{
		Replicas:      2,
		ContainerPort: 3000,
		ServicePort:   80,
		Labels:        map[string]string{},
	}

	if err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&values); err != nil && err != io.EOF {
		return fmt.Errorf("failed to parse input: %w", err)
	}

	// Add selector to labels
	maps.Copy(values.Labels, selector)

	resources := []flight.Resource{
		CreateDeployment(DeploymentConfig{
			Labels:   values.Labels,
			Replicas: 2,
			Port:     values.ContainerPort,
		}),
		CreateService(ServiceConfig{
			Labels:     values.Labels,
			Port:       values.ServicePort,
			TargetPort: values.ContainerPort,
		}),
	}

	return json.NewEncoder(os.Stdout).Encode(resources)
}

type DeploymentConfig struct {
	Labels   map[string]string
	Replicas int32
	Port     int32
}

func CreateDeployment(cfg DeploymentConfig) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.Identifier(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      release,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: selector,
			},
			Replicas: ptr.To(cfg.Replicas),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: cfg.Labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  release,
							Image: "nginx:latest",
							Ports: []corev1.ContainerPort{
								{ContainerPort: cfg.Port},
							},
						},
					},
				},
			},
		},
	}
}

type ServiceConfig struct {
	Labels     map[string]string
	Port       int32
	TargetPort int32
}

func CreateService(cfg ServiceConfig) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      release,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       cfg.Port,
					TargetPort: intstr.FromInt32(cfg.TargetPort),
				},
			},
		},
	}
}
