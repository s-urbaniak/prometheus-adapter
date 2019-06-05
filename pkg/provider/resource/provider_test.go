package resource

import (
	"encoding/json"
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	promapi "github.com/prometheus/client_golang/api/prometheus/v1"

	"github.com/prometheus/client_golang/api"

	metricsapiv1 "github.com/s-urbaniak/prometheus-adapter/pkg/apis/metrics/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

func finishPoll() (bool, error) {
	return true, nil
}

func cancelPoll(err error) (bool, error) {
	return true, err
}

func continuePoll() (bool, error) {
	return false, nil
}

func TestGetContainerMetrics(t *testing.T) {
	kubeConfigPath := "/home/sur/.kube/kind-config-kind"
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath}
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	config, err := loader.ClientConfig()
	if err != nil {
		t.Fatal(err)
	}

	spec := metricsapiv1.ResourceMetricsSpec{
		Rules: metricsapiv1.ResourceMetricsRules{

			CPU: metricsapiv1.ResourceMetricsRule{
				ContainerLabel: "container_name",
				Labels: map[string]metricsapiv1.GroupResource{
					"pod_name":  {Resource: "pod"},
					"namespace": {Resource: "namespace"},
					"node":      {Resource: "node"},
				},
				Queries: metricsapiv1.Queries{
					Pod:  `sum(rate(container_cpu_usage_seconds_total{container_name!="POD",container_name!="",pod_name="alertmanager-main-0",pod_name!="",namespace="monitoring"}[1m])) by (container_name,pod_name,namespace)`,
					Node: `sum(1 - rate(node_cpu_seconds_total{mode="idle"}[1m]) * on(namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{node="kind-control-plane"}) by (node)`,
				},
			},

			Memory: metricsapiv1.ResourceMetricsRule{
				ContainerLabel: "container_name",
				Labels: map[string]metricsapiv1.GroupResource{
					"pod_name":  {Resource: "pod"},
					"namespace": {Resource: "namespace"},
					"node":      {Resource: "node"},
				},
				Queries: metricsapiv1.Queries{
					Pod:  `sum(container_memory_working_set_bytes{pod_name="alertmanager-main-0",namespace="monitoring",container_name!="POD",container_name!="",pod_name!=""}) by (pod_name,namespace,container_name)`,
					Node: `sum(node:node_memory_bytes_total:sum{node="kind-control-plane"} - node:node_memory_bytes_available:sum{node="kind-control-plane"}) by (node)`,
				},
			},
		},
	}

	client, err := api.NewClient(api.Config{Address: "http://localhost:9090"})
	if err != nil {
		t.Fatal(err)
	}

	resources, err := restmapper.GetAPIGroupResources(discovery.NewDiscoveryClientForConfigOrDie(config))
	if err != nil {
		t.Fatal(err)
	}
	mapper := restmapper.NewDiscoveryRESTMapper(resources)

	p := NewProvider(
		mapper,
		promapi.NewAPI(client),
		spec,
	)

	_, metrics, err := p.GetContainerMetrics(types.NamespacedName{
		Name:      "prometheus-k8s-0",
		Namespace: "monitoring",
	})
	if err != nil {
		t.Fatal(err)
	}

	jsonBytes, err := json.MarshalIndent(metrics, "", "    ")
	if err != nil {
		t.Fatal(err)
	}

	yamlBytes, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(yamlBytes))
}
