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

type ServiceSpec struct {
	// +optional
	Name      string `json:"name"`
	Language  string `json:"language"`
	Framework string `json:"framework"`
}

type ServiceStatus struct {
	UID      string `json:"uid"`
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
}

// SimulationSpec defines the desired state of Simulation
type SimulationSpec struct {
	Name     string        `json:"name"`
	Services []ServiceSpec `json:"services"`
}

// SimulationStatus defines the observed state of Simulation
type SimulationStatus struct {
	ServicesStatus map[string]ServiceStatus `json:"services_status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Simulation is the Schema for the simulations API
type Simulation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SimulationSpec   `json:"spec,omitempty"`
	Status SimulationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SimulationList contains a list of Simulation
type SimulationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Simulation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Simulation{}, &SimulationList{})
}
