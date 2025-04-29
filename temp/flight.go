package temp

import (
	_ "embed"
	"fmt"

	"github.com/yokecd/yoke/pkg/helm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

//go:embed external-secrets-0.16.1.tgz
var archive []byte

// RenderChart renders the chart downloaded from https://charts.external-secrets.io/external-secrets
// Producing version: 0.16.1
func RenderChart(release, namespace string, values *Values) ([]*unstructured.Unstructured, error) {
	chart, err := helm.LoadChartFromZippedArchive(archive)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart from zipped archive: %w", err)
	}

	return chart.Render(release, namespace, values)
}
