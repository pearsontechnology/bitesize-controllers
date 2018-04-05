package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Policy struct {
    metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec               []PolicySpec `json:"spec"`
}

type PolicySpec struct {
    Path       string `json:"path"`
	Permission string `json:"permission"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PolicyList struct {
    metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items            []Policy `json:"items"`
}
