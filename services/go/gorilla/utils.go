package main

import (
	"bytes"
	"encoding/json"
	"github.com/tidwall/pretty"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
func callNextDestination(route json.RawMessage) (*Response, error) {
	var decodedPayload Route
	if err := json.Unmarshal(route, &decodedPayload); err != nil {
		return nil, err
	}

	var response *Response
	reqBody, err := json.Marshal(route)
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

// https://stackoverflow.com/a/40326580
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}