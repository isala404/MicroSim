package main

import (
	"bytes"
	"encoding/json"
	"github.com/tidwall/pretty"
	"io/ioutil"
	"log"
	"net/http"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf(
			"RemoteAddr=%s Method=%s Path=%s Body=%s",
			r.RemoteAddr,
			r.Method,
			r.RequestURI,
			pretty.Ugly(buf),
		)
		reader := ioutil.NopCloser(bytes.NewBuffer(buf))
		r.Body = reader
		next.ServeHTTP(w, r)
	})
}

// callNextDestination get the payload out ouf the request partially decoded it and send the raw data next Destination
func callNextDestination(payload json.RawMessage) (*Response, error) {
	var decodedPayload Payload
	if err := json.Unmarshal(payload, &decodedPayload); err != nil {
		return nil, err
	}

	var response *Response
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return response, err
	}

	resp, err := http.Post(decodedPayload.Designation, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}
	return response, nil
}
