package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	resourcemetrics "github.com/kubernetes-incubator/metrics-server/pkg/provider"
	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
	metricsv1 "github.com/s-urbaniak/prometheus-adapter/pkg/client/clientset/versioned/typed/metrics/v1"
	adapterql "github.com/s-urbaniak/prometheus-adapter/pkg/promql"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/metrics/pkg/apis/metrics"
)

var (
	_ resourcemetrics.MetricsProvider = (*provider)(nil)

	nsResource  = schema.GroupResource{Resource: "namespaces"}
	podResource = schema.GroupResource{Resource: "pods"}
)

type provider struct {
	promclient prom.API
	client     metricsv1.ResourceRuleInterface
	mapper     meta.RESTMapper
}

func NewProvider(resourceClient metricsv1.ResourceRuleInterface, mapper meta.RESTMapper, promclient prom.API) *provider {
	return &provider{
		client:     resourceClient,
		mapper:     mapper,
		promclient: promclient,
	}
}

func (p *provider) GetContainerMetrics(pods ...apitypes.NamespacedName) ([]resourcemetrics.TimeInfo, [][]metrics.ContainerMetrics, error) {
	if len(pods) == 0 {
		return nil, nil, nil
	}

	cpuRule, err := p.client.Get(string(corev1.ResourceCPU), metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("error loading cpu resource rules: %v", err)
	}

	cpuExpr, err := promql.ParseExpr(cpuRule.Spec.PodQuery)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing cpu expression: %v", err)
	}

	podsByNs := make(map[string][]string, len(pods))
	for _, pod := range pods {
		podsByNs[pod.Namespace] = append(podsByNs[pod.Namespace], pod.Name)
	}

	resources := make(map[string]string)
	for label, groupResource := range cpuRule.Spec.Labels {
		gvr, err := p.mapper.ResourceFor(schema.GroupResource{
			Group:    groupResource.Group,
			Resource: groupResource.Resource,
		}.WithVersion(""))

		if err != nil {
			return nil, nil, fmt.Errorf("error converting group resource: %v", err)
		}

		singular, err := p.mapper.ResourceSingularizer(gvr.Resource)
		if err != nil {
			return nil, nil, fmt.Errorf("error singularizing: %v", err)
		}

		resources[singular] = label
	}

	podResource, err := p.mapper.ResourceSingularizer(podResource.Resource)
	if err != nil {
		return nil, nil, fmt.Errorf("error singularizing pod: %v", err)
	}

	podLabel, ok := resources[podResource]
	if !ok {
		return nil, nil, errors.New("no pod label found")
	}

	nsResource, err := p.mapper.ResourceSingularizer(nsResource.Resource)
	if err != nil {
		return nil, nil, fmt.Errorf("error singularizing namespace: %v", err)
	}

	nsLabel, ok := resources[nsResource]
	if !ok {
		return nil, nil, errors.New("no namespace label found")
	}

	for ns, pods := range podsByNs {
		e := adapterql.NewEnforcer(
			&labels.Matcher{
				Name:  podLabel,
				Type:  labels.MatchRegexp,
				Value: strings.Join(pods, "|"),
			},
			&labels.Matcher{
				Name:  nsLabel,
				Type:  labels.MatchEqual,
				Value: ns,
			},
		)
		if err := e.EnforceNode(cpuExpr); err != nil {
			return nil, nil, fmt.Errorf("error enforcing cpu expression: %v", err)
		}
	}

	value, err := p.promclient.Query(context.Background(), cpuExpr.String(), time.Now())
	if value.Type() != prommodel.ValVector {
		return nil, nil, fmt.Errorf("invalid or empty value of non-vector type (%s) returned", value.Type())
	}

	return nil, nil, nil
}

func (p *provider) GetNodeMetrics(nodes ...string) ([]resourcemetrics.TimeInfo, []corev1.ResourceList, error) {
	return nil, nil, nil
}
