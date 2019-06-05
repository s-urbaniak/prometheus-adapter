package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	resourcemetrics "github.com/kubernetes-incubator/metrics-server/pkg/provider"
	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
	metricsv1 "github.com/s-urbaniak/prometheus-adapter/pkg/apis/metrics/v1"
	adapterql "github.com/s-urbaniak/prometheus-adapter/pkg/promql"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
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
	mapper     meta.RESTMapper
	spec       metricsv1.ResourceMetricsSpec
}

func NewProvider(mapper meta.RESTMapper, promclient prom.API, spec metricsv1.ResourceMetricsSpec) *provider {
	return &provider{
		mapper:     mapper,
		promclient: promclient,
		spec:       spec,
	}
}

type container struct {
	namespace, pod, container string
}

func (p *provider) executeQuery(r metricsv1.ResourceMetricsRule, podsByNamespace map[string][]string) (map[container]*prommodel.Sample, error) {
	expr, err := promql.ParseExpr(r.Queries.Pod)
	if err != nil {
		return nil, err
	}

	resources := make(map[string]string)
	for label, groupResource := range r.Labels {
		gvr, err := p.mapper.ResourceFor(schema.GroupResource{
			Group:    groupResource.Group,
			Resource: groupResource.Resource,
		}.WithVersion(""))

		if err != nil {
			return nil, fmt.Errorf("error converting group resource: %v", err)
		}

		singular, err := p.mapper.ResourceSingularizer(gvr.Resource)
		if err != nil {
			return nil, fmt.Errorf("error singularizing: %v", err)
		}

		resources[singular] = label
	}

	podResource, err := p.mapper.ResourceSingularizer(podResource.Resource)
	if err != nil {
		return nil, fmt.Errorf("error singularizing pod: %v", err)
	}

	podLabel, ok := resources[podResource]
	if !ok {
		return nil, errors.New("no pod label found in label spec")
	}

	nsResource, err := p.mapper.ResourceSingularizer(nsResource.Resource)
	if err != nil {
		return nil, fmt.Errorf("error singularizing namespace: %v", err)
	}

	nsLabel, ok := resources[nsResource]
	if !ok {
		return nil, errors.New("no namespace label found in label spec")
	}

	containerMetrics := make(map[container]*prommodel.Sample)

	for namespace, pods := range podsByNamespace {
		e := adapterql.NewEnforcer(
			&labels.Matcher{
				Name:  podLabel,
				Type:  labels.MatchRegexp,
				Value: strings.Join(pods, "|"),
			},
			&labels.Matcher{
				Name:  nsLabel,
				Type:  labels.MatchEqual,
				Value: namespace,
			},
		)

		if err := e.EnforceNode(expr); err != nil {
			return nil, fmt.Errorf("error enforcing expression: %v", err)
		}

		cpuValue, err := p.promclient.Query(context.Background(), expr.String(), time.Now())
		if err != nil {
			return nil, fmt.Errorf("error executing query: &%v", err)
		}

		cpuVector, ok := cpuValue.(prommodel.Vector)
		if !ok {
			return nil, fmt.Errorf("unsupported value type: %T", cpuValue)
		}

		earliestTs := prommodel.Latest

		for _, sample := range cpuVector {
			c := container{
				namespace: namespace,
				pod:       string(sample.Metric[prommodel.LabelName(podLabel)]),
				container: string(sample.Metric[prommodel.LabelName(r.ContainerLabel)]),
			}

			containerMetrics[c] = sample

			if sample.Timestamp.Before(earliestTs) {
				earliestTs = sample.Timestamp
			}
		}
	}

	return containerMetrics, nil
}

func (p *provider) GetContainerMetrics(pods ...apitypes.NamespacedName) ([]resourcemetrics.TimeInfo, [][]metrics.ContainerMetrics, error) {
	if len(pods) == 0 {
		return nil, nil, nil
	}

	podsByNs := make(map[string][]string, len(pods))
	for _, pod := range pods {
		podsByNs[pod.Namespace] = append(podsByNs[pod.Namespace], pod.Name)
	}

	cpuSamples, err := p.executeQuery(p.spec.Rules.CPU, podsByNs)
	if err != nil {
		return nil, nil, fmt.Errorf("error executing cpu query: %v", err)
	}

	memSamples, err := p.executeQuery(p.spec.Rules.Memory, podsByNs)
	if err != nil {
		return nil, nil, fmt.Errorf("error executing mem query: %v", err)
	}

	combined := make(map[container]*metrics.ContainerMetrics)
	for c, sample := range cpuSamples {
		combined[c] = &metrics.ContainerMetrics{
			Name:  c.container,
			Usage: corev1.ResourceList{},
		}

		combined[c].Usage[corev1.ResourceCPU] =
			*resource.NewMilliQuantity(int64(sample.Value*1000.0), resource.DecimalSI)
	}

	for c, sample := range memSamples {
		if _, present := combined[c]; !present {
			combined[c] = &metrics.ContainerMetrics{
				Name:  c.container,
				Usage: corev1.ResourceList{},
			}
		}

		combined[c].Usage[corev1.ResourceMemory] =
			*resource.NewMilliQuantity(int64(sample.Value*1000.0), resource.BinarySI)
	}

	podMetrics := make(map[apitypes.NamespacedName][]metrics.ContainerMetrics)
	for container, metrics := range combined {
		pod := apitypes.NamespacedName{
			Name:      container.pod,
			Namespace: container.namespace,
		}
		podMetrics[pod] = append(podMetrics[pod], *metrics)
	}

	var res [][]metrics.ContainerMetrics
	for _, pod := range pods {
		res = append(res, podMetrics[pod])
	}

	return nil, res, nil
}

func (p *provider) GetNodeMetrics(nodes ...string) ([]resourcemetrics.TimeInfo, []corev1.ResourceList, error) {
	return nil, nil, nil
}
