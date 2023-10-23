package connectors

import (
	camelv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Connector struct {
	metav1.TypeMeta   `json:",inline"            yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec ConnectorSpec `json:"spec,omitempty" yaml:"spec"`
}

type ConnectorSpec struct {
	Definition camelv1.Pipe `json:"definition" yaml:"definition"`

	// TODO investigate why unstructure.Unstructure does not work
	Resources []unstructured.Unstructured `json:"resources" yaml:"resources"`
}
