package v1

import (
	crdvalidation "github.com/ant31/crd-validation/pkg"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewResourceRuleCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: ResourceRules + "." + GroupName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: apiextensions.SchemeGroupVersion.Group + "/" + apiextensions.SchemeGroupVersion.Version,
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group:   GroupName,
			Version: GroupVersion.Version,
			Scope:   apiextensions.ClusterScoped,
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural: ResourceRules,
				Kind:   ResourceRulesKind,
			},
			Validation: crdvalidation.GetCustomResourceValidation(ResourceRuleSpecName, GetOpenAPIDefinitions),
		},
	}
}
