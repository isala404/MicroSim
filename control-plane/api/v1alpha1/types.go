package v1alpha1

import "encoding/json"

// +kubebuilder:object:generate=false
type FaultType struct {
	Type        string          `json:"type"`
	Args        json.RawMessage `json:"args"`
}

// +kubebuilder:object:generate=false
type Faults struct {
	Before []FaultType `json:"before"`
	After  []FaultType `json:"after"`
}

// +kubebuilder:object:generate=false
type Response struct {
	Service  string    `json:"service"`
	Address  string    `json:"address"`
	Errors   []string  `json:"errors"`
	Response *Response `json:"response"`
}

// +kubebuilder:object:generate=false
type Route struct {
	Designation string  `json:"designation"`
	Faults      Faults  `json:"faults"`
	Routes      []Route `json:"routes"`
}
