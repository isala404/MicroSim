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
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type FaultType struct {
	Type string          `json:"type"`
	Args json.RawMessage `json:"args"`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Probability int `json:"probability"`
}

type Faults struct {
	Before []FaultType `json:"before"`
	After  []FaultType `json:"after"`
}

type Payload struct {
	Designation string `json:"designation"`
	Faults      Faults `json:"faults"`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Probability int `json:"probability"`

	// TODO: Fix this
	// +optional
	// +nullable
	//Payload []Payload `json:"payload,omitempty"`
}

type Response struct {
	Service string   `json:"service"`
	Address string   `json:"address"`
	Errors  []string `json:"errors"`
	// +optional
	Response []Response `json:"response"`
}

// LoadGeneratorSpec defines the desired state of LoadGenerator
type LoadGeneratorSpec struct {
	Routes []Payload `json:"routes"`
	// +optional
	RequestCount int `json:"requestCount"`
	// +optional
	Timeout      time.Duration `json:"timeout"`
	BetweenDelay int           `json:"betweenDelay"`
}

// LoadGeneratorStatus defines the observed state of LoadGenerator
type LoadGeneratorStatus struct {
	// +kubebuilder:validation:Minimum=0
	DoneRequests int `json:"doneRequests"`
	// TODO: Fix this
	//Responses           []Response    `json:"responses"`
	AverageResponseTime time.Duration `json:"averageResponseTime"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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
