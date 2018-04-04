package crd

import (

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec               []PolicySpec `json:"spec"`
	Status             PolicyStatus `json:"status,omitempty"`
}

type PolicySpec struct {
    Path       string `json:"path"`
	Permission string `json:"permission"`
}

type PolicyStatus struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items            []Policy `json:"items"`
}
