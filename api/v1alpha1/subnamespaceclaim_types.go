// api/v1alpha1/subnamespaceclaim_types.go
package v1alpha1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleBindingTemplate defines a RoleBinding to be created in the sub-namespace
type RoleBindingTemplate struct {
	// Name of the RoleBinding
	Name string `json:"name"`
	// Subjects who will be given the role
	Subjects []rbacv1.Subject `json:"subjects"`
	// RoleRef specifies which role to bind to
	RoleRef rbacv1.RoleRef `json:"roleRef"`
}

// SubNamespaceClaimSpec defines the desired state of SubNamespaceClaim
type SubNamespaceClaimSpec struct {
	// Name is the name part of the namespace to be created
	// The actual namespace will be prefixed with the parent namespace
	Name string `json:"name"`
	// RoleBindings is a list of RoleBindings to create in the sub-namespace
	// +optional
	RoleBindings []RoleBindingTemplate `json:"roleBindings,omitempty"`
}

// SubNamespaceClaimStatus defines the observed state of SubNamespaceClaim
type SubNamespaceClaimStatus struct {
	// The full name of the created namespace (parent-ns + name)
	FullNamespace string `json:"fullNamespace,omitempty"`
	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Requested Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Full Namespace",type=string,JSONPath=`.status.fullNamespace`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// SubNamespaceClaim is the Schema for the subnamespaceclaims API
type SubNamespaceClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubNamespaceClaimSpec   `json:"spec,omitempty"`
	Status SubNamespaceClaimStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SubNamespaceClaimList contains a list of SubNamespaceClaim
type SubNamespaceClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SubNamespaceClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SubNamespaceClaim{}, &SubNamespaceClaimList{})
}
