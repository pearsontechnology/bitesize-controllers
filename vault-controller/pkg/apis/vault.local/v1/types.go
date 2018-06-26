package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type VaultPolicy struct {
    metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec               []VaultPolicySpec `json:"spec"`
}

type VaultPolicySpec struct {
    Path       string `json:"path"`
	Permission string `json:"permission"`
    TTL        string `json:"ttl"`
    Period     string `json:"period"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type VaultPolicyList struct {
    metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items            []VaultPolicy `json:"items"`
}
