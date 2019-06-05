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
	metricsclientv1 "github.com/s-urbaniak/prometheus-adapter/pkg/client/clientset/versioned/typed/metrics/v1"
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
	promclient    prom.API
	metricsclient metricsclientv1.ResourceRuleInterface
	mapper        meta.RESTMapper
}

func NewProvider(resourceClient metricsclientv1.ResourceRuleInterface, mapper meta.RESTMapper, promclient prom.API) *provider {
	return &provider{
		metricsclient: resourceClient,
		mapper:        mapper,
		promclient:    promclient,
	}
}

func (p *provider) GetContainerMetrics(pods ...apitypes.NamespacedName) ([]resourcemetrics.TimeInfo, [][]metrics.ContainerMetrics, error) {
	if len(pods) == 0 {
		return nil, nil, nil
	}

	cpuRule, err := p.metricsclient.Get(string(corev1.ResourceCPU), metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("error loading cpu resource rules: %v", err)
	}

	cpuExpr, err := promql.ParseExpr(cpuRule.Spec.PodQuery)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing cpu expression: %v", err)
	}

	memRule, err := p.metricsclient.Get(string(corev1.ResourceMemory), metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("error loading memory resource rules: %v", err)
	}

	memExpr, err := promql.ParseExpr(memRule.Spec.PodQuery)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing memory expression: %v", err)
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

	nsResource, err := p.mapper.ResourceSingularizer(nsResource.Resource)
	if err != nil {
		return nil, nil, fmt.Errorf("error singularizing namespace: %v", err)
	}

	nsLabel, ok := resources[nsResource]
	if !ok {
		return nil, nil, errors.New("no namespace label found in label spec")
	}

	podsByNs := make(map[string][]string, len(pods))
	for _, pod := range pods {
		podsByNs[pod.Namespace] = append(podsByNs[pod.Namespace], pod.Name)
	}

	podLabel, ok := resources[podResource]
	if !ok {
		return nil, nil, errors.New("no pod label found in label spec")
	}

	type container struct {
		namespace, pod, container string
	}

	containerMetrics := make(map[container]*metrics.ContainerMetrics)

	for namespace, pods := range podsByNs {
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

		if err := e.EnforceNode(cpuExpr); err != nil {
			return nil, nil, fmt.Errorf("error enforcing cpu expression: %v", err)
		}

		cpuValue, err := p.promclient.Query(context.Background(), cpuExpr.String(), time.Now())
		if err != nil {
			return nil, nil, fmt.Errorf("error executing mem query: &%v", err)
		}

		cpuVector, ok := cpuValue.(prommodel.Vector)
		if !ok {
			return nil, nil, fmt.Errorf("unsupported cpu value type: %T", cpuValue)
		}

		earliestTs := prommodel.Latest

		for _, sample := range cpuVector {
			c := container{
				namespace: namespace,
				pod:       string(sample.Metric[prommodel.LabelName(podLabel)]),
				container: string(sample.Metric[prommodel.LabelName(cpuRule.Spec.ContainerLabel)]),
			}

			if _, present := containerMetrics[c]; !present {
				containerMetrics[c] = &metrics.ContainerMetrics{
					Name:  c.container,
					Usage: corev1.ResourceList{},
				}
			}

			containerMetrics[c].Usage[corev1.ResourceCPU] =
				*resource.NewMilliQuantity(int64(sample.Value*1000.0), resource.DecimalSI)

			if sample.Timestamp.Before(earliestTs) {
				earliestTs = sample.Timestamp
			}
		}

		if err := e.EnforceNode(memExpr); err != nil {
			return nil, nil, fmt.Errorf("error enforcing memory expression: %v", err)
		}

		memValue, err := p.promclient.Query(context.Background(), memExpr.String(), time.Now())
		if err != nil {
			return nil, nil, fmt.Errorf("error executing memory query: &%v", err)
		}

		memVector, ok := memValue.(prommodel.Vector)
		if !ok {
			return nil, nil, fmt.Errorf("unsupported prometheus memory value type: %T", cpuValue)
		}

		for _, sample := range memVector {
			c := container{
				namespace: namespace,
				pod:       string(sample.Metric[prommodel.LabelName(podLabel)]),
				container: string(sample.Metric[prommodel.LabelName(memRule.Spec.ContainerLabel)]),
			}

			if _, present := containerMetrics[c]; !present {
				containerMetrics[c] = &metrics.ContainerMetrics{
					Name:  c.container,
					Usage: corev1.ResourceList{},
				}
			}

			containerMetrics[c].Usage[corev1.ResourceMemory] =
				*resource.NewMilliQuantity(int64(sample.Value*1000.0), resource.BinarySI)

			if sample.Timestamp.Before(earliestTs) {
				earliestTs = sample.Timestamp
			}
		}
	}

	podMetrics := make(map[apitypes.NamespacedName][]metrics.ContainerMetrics)
	for container, metrics := range containerMetrics {
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
