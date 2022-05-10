package handlers

import (
	"context"
	"net/http"
	"os"

	"github.com/egorovdmi/financify/foundation/database"
	"github.com/egorovdmi/financify/foundation/web"
	"github.com/jmoiron/sqlx"
)

type checkGroup struct {
	build string
	db    *sqlx.DB
}

func (cg checkGroup) readiness(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	status := "ok"
	statusCode := http.StatusOK

	err := database.StatusCheck(ctx, cg.db)
	if err != nil {
		status = "can't query database"
		statusCode = http.StatusInternalServerError
	}

	data := struct {
		Status string
	}{
		Status: status,
	}
	return web.Respond(ctx, rw, data, statusCode)
}

func (cg checkGroup) liveness(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {

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
		Build:     cg.build,
		Host:      host,
		Pod:       os.Getenv("KUBERNETES_PODNAME"),
		PodIP:     os.Getenv("KUBERNETES_NAMESPACE_POD_IP"),
		Node:      os.Getenv("KUBERNETES_NODENAME"),
		Namespace: os.Getenv("KUBERNETES_NAMESPACE"),
	}

	return web.Respond(ctx, rw, info, http.StatusOK)
}
