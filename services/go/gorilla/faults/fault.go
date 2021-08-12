package faults

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Fault interface {
	Run() error
}

type FaultType struct {
	Type string          `json:"type"`
	Args json.RawMessage `json:"args"`
}

type Faults []Fault

// UnmarshalJSON Map JSON objects native fault structs
func (f *Faults) UnmarshalJSON(data []byte) error {
	var faultTypes []FaultType
	err := json.Unmarshal(data, &faultTypes)
	if err != nil {
		return err
	}

	faults := make([]Fault, len(faultTypes))

	for i, faultType := range faultTypes {
		switch faultType.Type {
		case "latency":
			l := Latency{}
			if err := json.Unmarshal(faultType.Args, &l); err != nil {
				return err
			}
			faults[i] = l
		case "":
			return errors.New("fault type was not defined")
		default:
			return errors.New(fmt.Sprintf("fault type %s, is not implemented", faultType.Type))
		}
	}

	*f = faults
	return nil
}
