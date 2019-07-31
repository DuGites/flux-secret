package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StackOverDriveSecretSpec defines the desired state of StackOverDriveSecret
type StackOverDriveSecretSpec struct {
	EncryptedData map[string]string `json:"encryptedData"`
}

// StackOverDriveSecretStatus defines the observed state of StackOverDriveSecret
type StackOverDriveSecretStatus struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StackOverDriveSecret is the Schema for the stackoverdrivesecrets API
// +k8s:openapi-gen=true
type StackOverDriveSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StackOverDriveSecretSpec   `json:"spec,omitempty"`
	Status StackOverDriveSecretStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StackOverDriveSecretList contains a list of StackOverDriveSecret
type StackOverDriveSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StackOverDriveSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StackOverDriveSecret{}, &StackOverDriveSecretList{})
}
