package resource

import (
	resourcemetrics "github.com/kubernetes-incubator/metrics-server/pkg/provider"
	corev1 "k8s.io/api/core/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/metrics/pkg/apis/metrics"
)

var (
	_ resourcemetrics.MetricsProvider = (*provider)(nil)
)

type provider struct {
}

func NewProvider() *provider {
	return &provider{}
}

func (p *provider) GetContainerMetrics(pods ...apitypes.NamespacedName) ([]resourcemetrics.TimeInfo, [][]metrics.ContainerMetrics, error) {
	return nil, nil
}

func (p *provider) GetNodeMetrics(nodes ...string) ([]resourcemetrics.TimeInfo, []corev1.ResourceList, error) {
	return nil, nil
}
