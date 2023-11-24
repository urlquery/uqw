package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/hpcloud/tail"
	"github.com/urlquery/urlquery-api-go"
)

var cfg *AppConfig

func init() {
	var filename string
	var err error

	flag.StringVar(&filename, "config", "", "config file")
	flag.Parse()
	cfg, err = LoadConfig(filename)
	if err != nil {
		log.Fatal("load config", err)
	}

	if cfg.APIKey == "" || len(cfg.APIKey) != 32 {
		log.Fatal("API key needed")
	}
	urlquery.SetDefaultKey(cfg.APIKey)

}

func main() {

	// Start webhook server
	if cfg.Webhooks.Enabled {
		StartWebhookServer()
	}

	// Start submission workers
	for _, submitter := range cfg.Submit {
		if submitter.Enabled {
			go SubmitWorker(submitter)
		}
	}

	// TODO: add shutdown
	// listen for OS interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}

func SubmitWorker(cfg SubmitterSettings) {

	file, _ := os.OpenFile(cfg.File, os.O_RDWR|os.O_CREATE, 0644)
	file.Close()

	t, err := tail.TailFile(cfg.File, tail.Config{
		ReOpen:    true,
		MustExist: false,
		Poll:      true,
		Follow:    true,
		Logger:    tail.DiscardingLogger,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
	})
	if err != nil {
		log.Fatal(err)
	}

	for line := range t.Lines {
		url := urlquery.SubmitJob{
			Access: cfg.Settings.Access,
			Tags:   cfg.Settings.Tags,
		}

		url.Url = line.Text
		submission, err := urlquery.Submit(url)
		if err != nil {
			fmt.Println("submission failed", url.Url)
			continue
		}

		fmt.Println("submitted url", submission.QueueID, url.Url)
		if cfg.Output.Enabled {
			go GrabQueuedReport(submission.QueueID, cfg.Output)
		}
	}
}

// GrabQueuedReport waits for a report to finish
func GrabQueuedReport(queue_id string, output ReportOutput) {

	q, _ := urlquery.GetQueueStatus(queue_id)

	// TODO: Add timeout
	for q.Status != "done" && q.Status != "failed" {
		time.Sleep(4 * time.Second)
		q, _ = urlquery.GetQueueStatus(queue_id)
	}

	if q.Status == "done" {
		WriteReportData(q.ReportID, output)
	}

	if q.Status == "failed" {
		fmt.Println("Submission failed", q.Url)
	}

}

func WriteReportData(report_id string, evt ReportOutput) {
	path := GetOutputDir(evt.Path)

	if evt.Report {
		fmt.Println("getting report", report_id)

		report, _ := urlquery.GetReport(report_id)
		filename := fmt.Sprintf("%s/report_%s.json", path, report_id)
		os.WriteFile(filename, report.Bytes(), 0644)
	}

	if evt.Screenshot {
		fmt.Println("getting screenshot", report_id)

		screenshot, _ := urlquery.GetScreenshot(report_id)
		filename := fmt.Sprintf("%s/screenshot_%s.jpg", path, report_id)
		os.WriteFile(filename, screenshot, 0644)
	}

	if evt.DomainGraph {
		fmt.Println("getting domain graph")

		domain_graph, _ := urlquery.GetDomainGraph(report_id)
		filename := fmt.Sprintf("%s/domain_graph_%s.gif", path, report_id)
		os.WriteFile(filename, domain_graph, 0644)
	}
}

func GetOutputDir(path string) string {
	wd, _ := os.Getwd()
	output_path := strings.TrimSuffix(wd+"/"+path, "/")

	if strings.HasPrefix(path, "/") {
		// absolute path
		output_path = strings.TrimSuffix(path, "/")
	}
	return output_path
}
