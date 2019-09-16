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

	// Resources is a list of resources to apply the lock to
	//
	// For example:
	// 'pods' means pods.
	// 'pods/log' means the log subresource of pods.
	// '*' means all resources, but not subresources.
	// 'pods/*' means all subresources of pods.
	// '*/scale' means all scale subresources.
	// '*/*' means all resources and their subresources.
	//
	// If wildcard is present, the validation rule will ensure resources do not
	// overlap with each other.
	//
	// Depending on the enclosing object, subresources might not be allowed.
	Resources []string `json:"resources,omitempty"`

	// APIGroups is the API groups the resources belong to. '*' is all groups.
	// If '*' is present, the length of the slice must be one.
	APIGroups []string `json:"apiGroups,omitempty"`

	// APIVersions is the API versions the resources belong to. '*' is all versions.
	// If '*' is present, the length of the slice must be one.
	APIVersion []string `json:apiVersions,omitempty`

	// Operations which are permitted on the object, when empty blocks all CRUD
	Operations []string `json:"operations,omitempty"`

	// The message that will be returnes as a reason for locking the object
	Reason string `json:"reason,omitempty"`
}

// LockStatus reflect current state of the lockr
type LockStatus struct {

	// BlockedAttempts is the number of actions blocked by lock object
	// Blocked actions are those that violates the rules defined by the lock object
	BlockedAttempts int `json:blockedAttempts,omitempty`
}
