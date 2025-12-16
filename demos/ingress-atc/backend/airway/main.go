package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/yokecd/yoke/pkg/apis/v1alpha1"
	"github.com/yokecd/yoke/pkg/openapi"

	v1 "github.com/yokecd/examples/demos/ingress-atc/backend/v1"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	airway := v1alpha1.Airway{
		ObjectMeta: metav1.ObjectMeta{
			Name: "backends.examples.com",
		},
		Spec: v1alpha1.AirwaySpec{
			WasmURLs: v1alpha1.WasmURLs{
				Flight: "https://github.com/yokecd/examples/releases/download/latest/demos_ingress_atc.wasm.gz",
			},
			Template: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: v1.SchemeGroupVersion.Group,
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
	}

	return json.NewEncoder(os.Stdout).Encode(airway)
}
