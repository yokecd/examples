package main

import (
	"encoding/json"
	"flag"
	"os"
	"reflect"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/yokecd/yoke/pkg/apis/v1alpha1"
	"github.com/yokecd/yoke/pkg/openapi"

	v1 "github.com/yokecd/examples/demos/dynamic-mode/backend/v1"
)

func main() {
	subscription := flag.Bool("subscription", false, "use subscription instead of pure dynamic")

	flag.Parse()

	_ = json.NewEncoder(os.Stdout).Encode(v1alpha1.Airway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.APIVersion,
			Kind:       v1alpha1.KindAirway,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "backends.examples.com",
		},
		Spec: v1alpha1.AirwaySpec{
			WasmURLs: v1alpha1.WasmURLs{
				Flight: func() string {
					if *subscription {
						return "https://github.com/yokecd/examples/releases/download/latest/demos_dynamic_mode_v1_flight_subscription.wasm.gz"
					}
					return "https://github.com/yokecd/examples/releases/download/latest/demos_dynamic_mode_v1_flight.wasm.gz"
				}(),
			},
			Mode: func() v1alpha1.AirwayMode {
				if *subscription {
					return v1alpha1.AirwayModeSubscription
				}
				return v1alpha1.AirwayModeDynamic
			}(),
			ClusterAccess: true,
			Template: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: "examples.com",
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural:     "backends",
					Singular:   "backend",
					ShortNames: []string{"be"},
					Kind:       "Backend",
				},
				Scope: apiextensionsv1.NamespaceScoped,
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name:    "v1",
						Served:  true,
						Storage: true,
						Schema: &apiextensionsv1.CustomResourceValidation{
							OpenAPIV3Schema: openapi.SchemaFrom(reflect.TypeFor[v1.Backend]()),
						},
					},
				},
			},
		},
	})
}
