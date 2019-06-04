package naming

import (
	prom "github.com/prometheus/common/model"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type resourceConverter struct {
	resourceToLabel map[schema.GroupResource]prom.LabelName
}
