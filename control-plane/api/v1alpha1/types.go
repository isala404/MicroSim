package v1alpha1

import "encoding/json"

// +kubebuilder:object:generate=false
type Response struct {
	Service  string    `json:"service"`
	Address  string    `json:"address"`
	Errors   []string  `json:"errors"`
	Response *Response `json:"response"`
}

// +kubebuilder:object:generate=false
type Route struct {
	Designation string          `json:"designation"`
	Faults      json.RawMessage `json:"faults"`
	Routes      []Route         `json:"routes"`
}
