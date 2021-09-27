/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Responses struct {
	Request  string `json:"request"`
	Response string `json:"response"`
}

type SimulationRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// LoadGeneratorSpec defines the desired state of LoadGenerator
type LoadGeneratorSpec struct {
	Request string `json:"request"`
	// +optional
	// +kubebuilder:default=1
	Replicas      int           `json:"replicas"`
	SimulationRef SimulationRef `json:"simulationRef"`
	// +optional
	// +kubebuilder:validation:Minimum=0
	RequestCount *int `json:"requestCount"`
	// +optional
	Timeout      *metav1.Duration `json:"timeout"`
	BetweenDelay metav1.Duration  `json:"betweenDelay"`
}

// LoadGeneratorStatus defines the observed state of LoadGenerator
type LoadGeneratorStatus struct {
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0
	DoneRequests      int                  `json:"doneRequests"`
	Responses         map[string]Responses `json:"responses"`
	TotalResponseTime metav1.Duration      `json:"totalResponseTime"`
	Replicas          int                  `json:"replicas"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas

// LoadGenerator is the Schema for the loadgenerators API
type LoadGenerator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoadGeneratorSpec   `json:"spec,omitempty"`
	Status LoadGeneratorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LoadGeneratorList contains a list of LoadGenerator
type LoadGeneratorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoadGenerator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LoadGenerator{}, &LoadGeneratorList{})
}
