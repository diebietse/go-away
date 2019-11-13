package alertapi

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type AlertAPI struct {
}

func New() {
	r := mux.NewRouter()
	r.HandleFunc("/healthz", healthz).Methods("GET")
}

func (a *AlertAPI) alertHook(w http.ResponseWriter, r *http.Request) {

}

func healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Ok!")
}
