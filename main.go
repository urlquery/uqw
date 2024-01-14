package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"urlquery/public/uqw/webhook"

	"github.com/hpcloud/tail"
	"github.com/urlquery/urlquery-api-go"
)

var appcfg *AppConfig

func init() {
	var filename string
	var err error

	flag.StringVar(&filename, "config", "", "config file")
	flag.Parse()
	appcfg, err = LoadConfig(filename)
	if err != nil {
		log.Fatal("load config", err)
	}

	if appcfg.APIKey == "" || len(appcfg.APIKey) != 32 {
		log.Fatal("API key needed")
	}
	urlquery.SetDefaultKey(appcfg.APIKey)

	if appcfg.Webhooks.Enabled {
		fmt.Println("Using Webhooks")

		usr, _ := urlquery.GetUser()
		if err == nil {
			fmt.Println("  NB: Make sure the configured Webhook is correct (requests originate from urlquery.net)")
			fmt.Println("  Webhook URL:", usr.Notify.Webhook.URL)

			if usr.Notify.Webhook.Enabled == false {
				fmt.Println("  WARNING - Webhooks is not enabled")
			}
		}
	}

}

func main() {
	var srv *webhook.WebhookServer

	// Start webhook server
	if appcfg.Webhooks.Enabled {
		fmt.Println("Starting Webhook server:", appcfg.Webhooks.Listen)
		srv = webhook.CreateWebhookServer(appcfg.Webhooks.Listen)

		go func() {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}()

		srv.RegisterCallbackReportCompleted(appcfg.Webhooks.Reports.Submitted.WriteReportData)
		srv.RegisterCallbackAlertedReport(appcfg.Webhooks.Reports.Alerted.WriteReportData)
	}

	// Start submission workers
	for _, submitter := range appcfg.Submit {
		if submitter.Enabled {
			go SubmitWorker(submitter)
		}
	}

	waitForQuitSignal()
	err := shutdownHttpServer(srv.Server, 60)
	if err != nil {
		fmt.Println("Shutdown timedout")
		os.Exit(1)
	}

}

// waitForQuitSignal waits for CRTL-C, forces exit if its pressed 2 more times
func waitForQuitSignal() {
	// listen for OS interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	go func() {
		for force_quit := 2; force_quit != 0; force_quit-- {
			fmt.Printf("Hit Ctrl-C %d more times to force quit..\n", force_quit)
			signal.Notify(quit, os.Interrupt)
			<-quit
		}
		os.Exit(1)
	}()
}

func shutdownHttpServer(srv *http.Server, timeout_seconds uint) error {
	if srv == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout_seconds)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
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
			go cfg.Output.GrabQueuedReport(submission.QueueID)
		}
	}
}
