package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/MrSupiri/MicroSim/service/gorilla/faults"
	"github.com/gorilla/mux"
	"github.com/tidwall/pretty"
	"log"
	"net/http"
)

var serviceName string
var port string

type Route struct {
	Designation string `json:"designation,omitempty"`
	Faults      struct {
		Before faults.Faults `json:"before,omitempty"`
		After  faults.Faults `json:"after,omitempty"`
	} `json:"faults"`
	Routes []json.RawMessage `json:"routes"`
}

type Response struct {
	Service  string      `json:"service"`
	Address  string      `json:"address"`
	Errors   []string    `json:"errors"`
	Response []*Response `json:"response"`
}

func main() {
	flag.StringVar(&serviceName, "service-name", "Undefined", "The name set on the response")
	flag.StringVar(&port, "addr", ":8080", "The address the web server will bind to")
	flag.Parse()

	// override of if ENV is present
	serviceName = getEnv("SERVICE_NAME", serviceName)

	r := mux.NewRouter()
	r.HandleFunc("/", handler).Methods(http.MethodPost)
	r.Use(mux.CORSMethodMiddleware(r))
	r.Use(loggingMiddleware)

	fmt.Printf("service: %s, started on %s\n", serviceName, port)
	panic(http.ListenAndServe(port, r))
}

func handler(w http.ResponseWriter, r *http.Request) {
	var payload Route
	res := Response{Service: serviceName, Errors: []string{}, Response: []*Response{}}
	reqID := r.Header.Get("X-Request-ID")
	w.Header().Set("content-type", "application/json")

	// Get the request payload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res.Errors = append(res.Errors, err.Error())
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// Set the incoming request designation as the address for this service
	res.Address = payload.Designation

	// Run fault faults
	for _, fault := range payload.Faults.Before {
		if err := fault.Run(); err != nil {
			res.Errors = append(res.Errors, err.Error())
		}
	}

	res.Response = make([]*Response, len(payload.Routes))
	// Forward the request to next service if the destination is defined
	for i, route := range payload.Routes {
		destRes, err := callNextDestination(route, reqID)
		if err != nil {
			res.Errors = append(res.Errors, err.Error())
			log.Println("error while forwarding request", err)
		}
		res.Response[i] = destRes
	}
	// Run post faults
	for _, fault := range payload.Faults.After {
		if err := fault.Run(); err != nil {
			res.Errors = append(res.Errors, err.Error())
		}
	}
	if resEn, err := json.Marshal(res); err == nil {
		log.Printf(
			"RequestID=%s, Response=%s",
			reqID,
			pretty.Ugly(resEn),
		)
	}
	w.WriteHeader(http.StatusOK)
	// Return the response to calling service
	_ = json.NewEncoder(w).Encode(res)
	r.Body.Close()
}
