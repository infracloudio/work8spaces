package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WorkspaceUserSpec defines the desired state of WorkspaceUser
type WorkspaceUserSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// TODO(bhavin192): do we need this to be minimum 1
	// +kubebuilder:validation:MinItems=1
	Workspaces []WUWorkspaceItem `json:"workspaces,omitempty"`
}

// WUWorkspaceItem defines name of Workspace and role of user in it.
type WUWorkspaceItem struct {
	// Name defines the name of a workspace
	Name string `json:"name,omitempty"`
	// Role defines the name of role granted to the user
	Role string `json:"role,omitempty"`
}

// WorkspaceUserStatus defines the observed state of WorkspaceUser
type WorkspaceUserStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkspaceUser is the Schema for the workspaceusers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=workspaceusers,scope=Cluster
type WorkspaceUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceUserSpec   `json:"spec,omitempty"`
	Status WorkspaceUserStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkspaceUserList contains a list of WorkspaceUser
type WorkspaceUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkspaceUser{}, &WorkspaceUserList{})
}
