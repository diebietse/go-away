package utils

import (
	"testing"

	"github.com/diebietse/go-away/firealert"
	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
)

func TestValidInputFiring(t *testing.T) {
	name := "test"
	severity := "ERROR"
	status := "firing"
	summary := "this is a summary"

	alert, err := AlertExtract(generateAlert(name, severity, status, summary))
	assert.NoError(t, err, "The valid alert does not parse")
	assert.Equal(t, name, alert.ID, "ID mismatch")
	assert.Equal(t, name, alert.Title, "Title mismatch")
	assert.Equal(t, firealert.LevelError, alert.Level, "Alert level mismatch")
	assert.Equal(t, firealert.StateTriggered, alert.State, "Invalid alert state")
	assert.Contains(t, alert.Body, summary, "Summary not in body")
	assert.Contains(t, alert.Body, severity, "Severity not in body")
}

func TestValidInputResolved(t *testing.T) {
	name := "test"
	severity := "DEBUG"
	status := "resolved"
	summary := "this is a summary"

	alert, err := AlertExtract(generateAlert(name, severity, status, summary))
	assert.NoError(t, err, "The valid alert does not parse")
	assert.Equal(t, name, alert.ID, "ID mismatch")
	assert.Equal(t, name, alert.Title, "Title mismatch")
	assert.Equal(t, firealert.LevelDebug, alert.Level, "Alert level mismatch")
	assert.Equal(t, firealert.StateResolved, alert.State, "Invalid alert state")
	assert.Contains(t, alert.Body, summary, "Summary not in body")
	assert.Contains(t, alert.Body, severity, "Severity not in body")
}

func TestUnknownStates(t *testing.T) {
	alert, err := AlertExtract(generateAlert("name", "severity", "status", "summary"))
	assert.NoError(t, err, "The valid alert does not parse")
	assert.Equal(t, firealert.LevelUnknown, alert.Level, "Alert level mismatch")
	assert.Equal(t, firealert.StateUnknown, alert.State, "Alert level mismatch")
}

func TestNoSeverity(t *testing.T) {
	data := &template.Data{CommonLabels: template.KV{}}
	_, err := AlertExtract(data)
	assert.Error(t, err)
}

func TestNoName(t *testing.T) {
	data := &template.Data{CommonLabels: template.KV{"severity": "HIGH"}}
	_, err := AlertExtract(data)
	assert.Error(t, err)
}

func generateAlert(name, severity, status, summary string) *template.Data {
	return &template.Data{
		Status: status,
		CommonLabels: template.KV{
			"alertname": name,
			"severity":  severity,
		},
		CommonAnnotations: template.KV{
			"summary": summary,
		},
	}
}
