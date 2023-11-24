package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type webhookServer struct {
	router *mux.Router
	cfg    WebhooksSettings
}

func StartWebhookServer() {
	srv := CreateWebhookServer(cfg.Webhooks)
	http.Handle("/", srv.router)

	// Start HTTP server
	web := &http.Server{
		Addr:         cfg.Webhooks.Listen,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	go func() {
		if err := web.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
}

func CreateWebhookServer(cfg WebhooksSettings) *webhookServer {
	var server webhookServer

	server.cfg = cfg
	server.router = mux.NewRouter()
	server.router.HandleFunc("/report/", server.webhookReportHandler).Methods("GET")

	return &server
}

func (srv webhookServer) webhookReportHandler(w http.ResponseWriter, r *http.Request) {

	report_id := r.URL.Query().Get("report_id")
	event := r.URL.Query().Get("event")

	fmt.Println(r.RequestURI)

	switch event {
	case "completed":
		go WriteReportData(report_id, srv.cfg.Reports.Submitted)

	case "alerted":
		go WriteReportData(report_id, srv.cfg.Reports.Alerted)
	}

	w.WriteHeader(http.StatusOK)
}
