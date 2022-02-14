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
			"RequestID=%s, RemoteAddr=%s, Method=%s, Path=%s, Body=%s",
			r.Header.Get("X-Request-ID"),
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
func callNextDestination(route json.RawMessage, reqID string) (*Response, error) {
	var decodedPayload Route
	if err := json.Unmarshal(route, &decodedPayload); err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(route)
	if err != nil {
		return nil, err
	}
	log.Printf("RequestID=%s, Calling Next Destination, Designation=%s, Body=%s", reqID, decodedPayload.Designation, pretty.Ugly(reqBody))
	var response *Response

	client := http.Client{}
	req, err := http.NewRequest("POST", decodedPayload.Designation, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"Content-Type": []string{"application/json"},
		"X-Request-ID": []string{reqID},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
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
