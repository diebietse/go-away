package main

import (
	"context"
	"net/http"
	"time"

	"github.com/diebietse/go-away/alertapi"
	"github.com/diebietse/go-away/firealert"
	"github.com/diebietse/go-away/utils"
	flags "github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"

	"golang.org/x/time/rate"
)

type Config struct {
	ListenAddress  string  `short:"a" long:"listen-address" description:"Listen address of the API server" value-name:"LISTEN_ADDRESS"`
	BurstLimit     int     `short:"b" long:"burst-limit" description:"Burst limit for the API server" value-name:"BURST_LIMIT"`
	RateLimit      float64 `short:"r" long:"rate-limit" description:"Rate limit for the API server" value-name:"RATE_LIMIT"`
	TimeoutSeconds int     `short:"t" long:"timeout-seconds" description:"Timeout in seconds to google API" value-name:"TIMEOUTS"`
	GoogleCreds    string  `short:"g" long:"google-creds" description:"The file with the google credentials" value-name:"GOOGLE_CREDS"`
	Level          string  `short:"l" long:"report-level" description:"Alert level to cause popup notification" value-name:"REPORT_LEVEL"`
}

func main() {
	c := Config{}
	if _, err := flags.Parse(&c); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			return
		} else {
			log.Fatalf("Could not parse arguments: %v", err)
		}
	}
	api := alertapi.New(rate.Limit(c.RateLimit), c.BurstLimit, c.TimeoutSeconds)
	alert, err := firealert.New(c.GoogleCreds, firealert.AlertLevel(1))
	if err != nil {
		log.Fatalf("Could not start firebase: %v", err)
	}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	go func() {
		for a := range api.Alerts() {
			alrt, err := utils.AlertExtract(a)
			if err != nil {
				log.Errorf("Could not parse alert: %+v", err)
				continue
			}
			if err = alert.SendAlert(ctx, *alrt); err != nil {
				log.Errorf("Could not send alert: %+v", err)
			}
		}
	}()

	log.Printf("listening on: %v", c.ListenAddress)
	log.Fatal(http.ListenAndServe(c.ListenAddress, api))
}
