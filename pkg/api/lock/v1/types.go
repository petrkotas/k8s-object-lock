package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Lock is a specification for a Foo resource
type Lock struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LockSpec   `json:"spec,omitempty"`
	Status LockStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LockList is a list of Foo resources
type LockList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Lock `json:"items"`
}

// LockSpec defines what operations are locked for an object
type LockSpec struct {
	// Operations which are permitted on the object, when empty blocks all CRUD
	Operations []string `json:"operations,omitempty"`
	// SubResources which are permitted in the object, when empty block all
	SubResources []string `json:"subresources,omitempty"`
	// The message that will be returnes as a reason for locking the object
	Reason string `json:"reason,omitempty"`
}

// LockStatus reflect current state of the lockr
type LockStatus struct {
}
