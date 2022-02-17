package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

type check struct {
	log *log.Logger
}

func (c check) readiness(rw http.ResponseWriter, r *http.Request) {
	status := struct {
		Status string
	}{
		Status: "OK",
	}
	json.NewEncoder(rw).Encode(status)
}
