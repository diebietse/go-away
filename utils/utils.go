package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/diebietse/go-away/firealert"
	"github.com/prometheus/alertmanager/template"
)

func AlertExtract(data *template.Data) (*firealert.Alert, error) {
	alert := &firealert.Alert{}
	switch data.Status {
	case "firing":
		alert.State = firealert.StateTriggered
	case "resolved":
		alert.State = firealert.StateResolved
	default:
		alert.State = firealert.StateUnknown
	}

	severity, ok := data.CommonLabels["severity"]
	if !ok {
		return nil, errors.New("could not find severity")
	}
	switch severity {
	case firealert.AlertStrings[firealert.LevelDebug]:
		alert.Level = firealert.LevelDebug
	case firealert.AlertStrings[firealert.LevelInfo]:
		alert.Level = firealert.LevelInfo
	case firealert.AlertStrings[firealert.LevelWarning]:
		alert.Level = firealert.LevelWarning
	case firealert.AlertStrings[firealert.LevelError]:
		alert.Level = firealert.LevelError
	default:
		alert.Level = firealert.LevelUnknown
	}

	alertname, ok := data.CommonLabels["alertname"]
	if !ok {
		return nil, errors.New("could not find alert name")
	}
	alert.ID = alertname
	alert.Title = alertname
	alert.Time = time.Now()

	summary := data.CommonAnnotations["summary"]
	alert.Body = fmt.Sprintf("%s: %s", severity, summary)

	return alert, nil
}
