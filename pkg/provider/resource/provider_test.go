package resource

import (
	"encoding/json"
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	metricsapiv1 "github.com/s-urbaniak/prometheus-adapter/pkg/apis/metrics/v1"
	metricsclientv1 "github.com/s-urbaniak/prometheus-adapter/pkg/client/clientset/versioned/typed/metrics/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func TestGetContainerMetrics(t *testing.T) {
	kubeConfigPath := "/home/sur/.kube/kind-config-kind"
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath}
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	config, err := loader.ClientConfig()
	if err != nil {
		t.Fatalf("unable to construct  auth configuration from %q for connecting to Prometheus: %v", kubeConfigPath, err)
	}

	apiclient := apiextensions.NewForConfigOrDie(config)

	crd := metricsapiv1.NewResourceRuleCRD()
	err = apiclient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("resourcerules.metrics.prometheus.io", &metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		t.Fatal(err)
	}

	_, err = apiclient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatal(err)
	}

	rr := &metricsapiv1.ResourceRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: metricsapiv1.GroupVersion.Version,
			Kind:       metricsapiv1.ResourceRulesKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: string(v1.ResourceCPU),
		},
		Spec: metricsapiv1.ResourceRuleSpec{
			NodeQuery: `sum(1 - rate(node_cpu_seconds_total{mode="idle"}[1m]) * on(namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{node="kind-control-plane"}) by (node)`,
			PodQuery:  `sum(rate(container_cpu_usage_seconds_total{container_name="alertmanager",container_name!="POD",container_name!="",pod_name="alertmanager-main-0",pod_name!="",namespace="monitoring"}[1m])) by (container_name,pod_name,namespace)`,
			Labels: map[string]metricsapiv1.GroupResource{
				"pod_name":  {Resource: "pod"},
				"namespace": {Resource: "namespace"},
				"node":      {Resource: "node"},
			},
			ContainerLabel: "container_name",
		},
	}

	jsonBytes, err := json.MarshalIndent(rr, "", "    ")
	if err != nil {
		t.Fatal(err)
	}

	yamlBytes, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(yamlBytes))

	metricsclient := metricsclientv1.NewForConfigOrDie(config)
	rr, err = metricsclient.ResourceRules().Create(rr)
	if err != nil {
		t.Fatal(err)
	}
}
