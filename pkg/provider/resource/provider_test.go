package resource

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	promapi "github.com/prometheus/client_golang/api/prometheus/v1"

	"github.com/prometheus/client_golang/api"

	"k8s.io/apimachinery/pkg/util/wait"

	v1 "k8s.io/api/core/v1"

	metricsapiv1 "github.com/s-urbaniak/prometheus-adapter/pkg/apis/metrics/v1"
	metricsclientv1 "github.com/s-urbaniak/prometheus-adapter/pkg/client/clientset/versioned/typed/metrics/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	apiclient := apiextensions.NewForConfigOrDie(config)
	crd := metricsapiv1.NewResourceRuleCRD()
	err = apiclient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(
		metricsapiv1.ResourceRules+"."+metricsapiv1.GroupName,
		&metav1.DeleteOptions{},
	)
	if err != nil && !apierrors.IsNotFound(err) {
		t.Fatal(err)
	}

	err = wait.Poll(500*time.Millisecond, 10*time.Second, func() (bool, error) {
		_, err := apiclient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(
			metricsapiv1.ResourceRules+"."+metricsapiv1.GroupName,
			metav1.GetOptions{},
		)
		if apierrors.IsNotFound(err) {
			return finishPoll()
		}
		if err != nil {
			return cancelPoll(err)
		}
		return continuePoll()
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = apiclient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatal(err)
	}

	err = wait.Poll(500*time.Millisecond, 10*time.Second, func() (bool, error) {
		_, err := apiclient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(
			metricsapiv1.ResourceRules+"."+metricsapiv1.GroupName,
			metav1.GetOptions{},
		)
		if apierrors.IsNotFound(err) {
			return continuePoll()
		}
		if err != nil {
			return cancelPoll(err)
		}
		return finishPoll()
	})
	if err != nil {
		t.Fatal(err)
	}

	metricsclient := metricsclientv1.NewForConfigOrDie(config)

	cpuRule := &metricsapiv1.ResourceRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: metricsapiv1.GroupVersion.Version,
			Kind:       metricsapiv1.ResourceRulesKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: string(v1.ResourceCPU),
		},
		Spec: metricsapiv1.ResourceRuleSpec{
			NodeQuery: `sum(1 - rate(node_cpu_seconds_total{mode="idle"}[1m]) * on(namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{node="kind-control-plane"}) by (node)`,
			PodQuery:  `sum(rate(container_cpu_usage_seconds_total{container_name!="POD",container_name!="",pod_name="alertmanager-main-0",pod_name!="",namespace="monitoring"}[1m])) by (container_name,pod_name,namespace)`,
			Labels: map[string]metricsapiv1.GroupResource{
				"pod_name":  {Resource: "pod"},
				"namespace": {Resource: "namespace"},
				"node":      {Resource: "node"},
			},
			ContainerLabel: "container_name",
		},
	}

	cpuRule, err = metricsclient.ResourceRules().Create(cpuRule)
	if err != nil {
		t.Fatal(err)
	}

	memRule := &metricsapiv1.ResourceRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: metricsapiv1.GroupVersion.Version,
			Kind:       metricsapiv1.ResourceRulesKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: string(v1.ResourceMemory),
		},
		Spec: metricsapiv1.ResourceRuleSpec{
			NodeQuery: `sum(node:node_memory_bytes_total:sum{node="kind-control-plane"} - node:node_memory_bytes_available:sum{node="kind-control-plane"}) by (node)`,
			PodQuery:  `sum(container_memory_working_set_bytes{pod_name="alertmanager-main-0",namespace="monitoring",container_name!="POD",container_name!="",pod_name!=""}) by (pod_name,namespace,container_name)`,
			Labels: map[string]metricsapiv1.GroupResource{
				"pod_name":  {Resource: "pod"},
				"namespace": {Resource: "namespace"},
				"node":      {Resource: "node"},
			},
			ContainerLabel: "container_name",
		},
	}

	memRule, err = metricsclient.ResourceRules().Create(memRule)
	if err != nil {
		t.Fatal(err)
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
		metricsclient.ResourceRules(),
		mapper,
		promapi.NewAPI(client),
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
