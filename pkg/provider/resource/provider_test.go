package resource

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/s-urbaniak/prometheus-adapter/pkg/client/clientset/versioned/typed/metrics/v1"
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

	client, err := v1.NewForConfig(config)
	rules, err := client.ResourceRules().List(metav1.ListOptions{})
	t.Log(rules, err)
}
