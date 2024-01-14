package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urlquery/urlquery-api-go"
	yaml "gopkg.in/yaml.v3"
)

type AppConfig struct {
	APIKey string `yaml:"apikey"`

	Webhooks WebhooksSettings    `yaml:"webhooks"`
	Submit   []SubmitterSettings `yaml:"submit"`
}

type WebhooksSettings struct {
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`

	Reports struct {
		Alerted   ReportOutput `yaml:"alerted"`
		Submitted ReportOutput `yaml:"submitted"`
	} `yaml:"reports"`
}

type SubmitterSettings struct {
	File     string `yaml:"file"`
	Enabled  bool   `yaml:"enabled"`
	Settings struct {
		Access string            `yaml:"access"`
		Tags   []string          `yaml:"tags"`
		Meta   map[string]string `yaml:"meta"`
	} `yaml:"settings"`
	Output ReportOutput `yaml:"output"`
}

type ReportOutput struct {
	Enabled     bool   `yaml:"enabled"`
	Path        string `yaml:"path"`
	Report      bool   `yaml:"report"`
	Screenshot  bool   `yaml:"screenshot"`
	DomainGraph bool   `yaml:"domain_graph"`
}

func LoadConfig(filename string) (*AppConfig, error) {
	var cfg AppConfig

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// GrabQueuedReport waits for a report to finish
func (output ReportOutput) GrabQueuedReport(queue_id string) {

	q, _ := urlquery.GetQueueStatus(queue_id)

	// TODO: Add timeout
	for q.Status != "done" && q.Status != "failed" {
		time.Sleep(4 * time.Second)
		q, _ = urlquery.GetQueueStatus(queue_id)
	}

	if q.Status == "done" {
		output.WriteReportData(q.ReportID)
	}

	if q.Status == "failed" {
		fmt.Println("Submission failed", q.Url)
	}

}

func (r ReportOutput) WriteReportData(report_id string) {
	path := GetOutputDir(r.Path)

	if r.Report {
		fmt.Println("getting report", report_id)

		report, _ := urlquery.GetReport(report_id)
		filename := fmt.Sprintf("%s/report_%s.json", path, report_id)
		os.WriteFile(filename, report.Bytes(), 0644)
	}

	if r.Screenshot {
		fmt.Println("getting screenshot", report_id)

		screenshot, _ := urlquery.GetScreenshot(report_id)
		filename := fmt.Sprintf("%s/screenshot_%s.jpg", path, report_id)
		os.WriteFile(filename, screenshot, 0644)
	}

	if r.DomainGraph {
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
