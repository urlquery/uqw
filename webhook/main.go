package webhook

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type WebhookServer struct {
	*http.Server

	funcReportCompleted []func(string)
	funcAlertedReport   []func(string)
}

func (srv *WebhookServer) RegisterCallbackReportCompleted(f func(string)) {
	srv.funcReportCompleted = append(srv.funcReportCompleted, f)
}
func (srv *WebhookServer) RegisterCallbackAlertedReport(f func(string)) {
	srv.funcAlertedReport = append(srv.funcAlertedReport, f)
}

func CreateWebhookServer(listen string) *WebhookServer {
	router := mux.NewRouter()
	server := WebhookServer{
		Server: &http.Server{
			Addr:         listen,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			Handler:      router,
		},
	}

	router.HandleFunc("/report/", server.webhookReportHandler).Methods("GET")

	return &server
}

func (srv *WebhookServer) webhookReportHandler(w http.ResponseWriter, r *http.Request) {
	report_id := r.URL.Query().Get("report_id")
	event := r.URL.Query().Get("event")

	fmt.Println(r.RequestURI)

	switch event {
	case "completed":
		for _, f := range srv.funcReportCompleted {
			go f(report_id)
		}

	case "alerted":
		for _, f := range srv.funcAlertedReport {
			go f(report_id)
		}

	}

	w.WriteHeader(http.StatusOK)
}
