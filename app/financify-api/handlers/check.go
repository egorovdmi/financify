package handlers

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/egorovdmi/financify/foundation/web"
)

type check struct {
	log   *log.Logger
	build string
}

func (c check) readiness(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	status := struct {
		Status string
	}{
		Status: "OK",
	}
	return web.Respond(ctx, rw, status, http.StatusOK)
}

func (c check) liveness(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {

	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	info := struct {
		Status    string `json:"status,omitempty"`
		Build     string `json:"build,omitempty"`
		Host      string `json:"host,omitempty"`
		Pod       string `json:"pod,omitempty"`
		PodIP     string `json:"podIP,omitempty"`
		Node      string `json:"node,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	}{
		Status:    "up",
		Build:     c.build,
		Host:      host,
		Pod:       os.Getenv("KUBERNETES_PODNAME"),
		PodIP:     os.Getenv("KUBERNETES_NAMESPACE_POD_IP"),
		Node:      os.Getenv("KUBERNETES_NODENAME"),
		Namespace: os.Getenv("KUBERNETES_NAMESPACE"),
	}

	return web.Respond(ctx, rw, info, http.StatusOK)
}
