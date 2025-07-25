// Package v1beta1 contains the input type for this Function
// +kubebuilder:object:generate=true
// +groupName=template.fn.crossplane.io
// +versionName=v1beta1
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Input is the input type for the XFleetManager function.
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=crossplane
type Input struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InputSpec `json:"spec"`
}

type InputSpec struct {
	Parameters FleetManagerParameters `json:"parameters"`
}

type FleetManagerParameters struct {
	FleetName     string            `json:"fleetName"`
	Region        string            `json:"region"`
	InstanceCount int               `json:"instanceCount"`
	Environment   string            `json:"environment"`
	Tags          map[string]string `json:"tags,omitempty"`
}
