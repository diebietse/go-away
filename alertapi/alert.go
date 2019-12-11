package alertapi

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/alertmanager/template"
	log "github.com/sirupsen/logrus"

	"golang.org/x/time/rate"
)

const (
	healthzPath      = "/healthz"
	alertmanagerPath = "/alertmanager"
)

const okBody = "Ok!"

type AlertAPI struct {
	*mux.Router
	limiter *rate.Limiter
	alerts  chan *template.Data
	timeout time.Duration
}

type responseJSON struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func New(r rate.Limit, b int, timeoutSec int) *AlertAPI {
	alert := &AlertAPI{
		Router:  mux.NewRouter(),
		limiter: rate.NewLimiter(r, b),
		alerts:  make(chan *template.Data, 1),
		timeout: time.Second * time.Duration(timeoutSec),
	}

	alert.HandleFunc(healthzPath, healthz).Methods("GET")
	alert.HandleFunc(alertmanagerPath, alert.rateLimitHook).Methods("POST")

	return alert
}

func (a *AlertAPI) Alerts() chan *template.Data {
	return a.alerts
}

func (a *AlertAPI) rateLimitHook(w http.ResponseWriter, r *http.Request) {
	if !a.limiter.Allow() {
		asJSON(w, http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
		return
	}
	ctx := context.Background()
	ctx, cancle := context.WithTimeout(ctx, a.timeout)
	defer cancle()
	a.alertHook(ctx, w, r)
}

func (a *AlertAPI) alertHook(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		asJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	alerts := &template.Data{}

	if err = json.Unmarshal(data, alerts); err != nil {
		log.Errorf("Invalid alert: %v", err)
		asJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	select {
	case <-ctx.Done():
		log.Printf("Timeout")
		asJSON(w, http.StatusRequestTimeout, http.StatusText(http.StatusRequestTimeout))
		return
	case a.alerts <- alerts:
		asJSON(w, http.StatusOK, "success")
	}
}

func asJSON(w http.ResponseWriter, status int, message string) {
	response := responseJSON{
		Status:  status,
		Message: message,
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Errorf("Could create JSON response: %v", err)
		return
	}

	w.WriteHeader(status)
	_, err = w.Write(data)
	if err != nil {
		log.Errorf("Could write response: %v", err)
	}
}

func healthz(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(okBody))
	if err != nil {
		log.Errorf("Could write response: %v", err)
	}
}
