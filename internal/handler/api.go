package handler

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/tp_security/internal/config"
)

type controller interface {
	GetOne(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Repeat(w http.ResponseWriter, r *http.Request)
	Scan(w http.ResponseWriter, r *http.Request)
}

func NewApi(cfg *config.Config, contr controller) http.Handler {
	rout := mux.NewRouter()
	rout.HandleFunc("/api/requests/{id}", contr.GetOne).Methods("GET")
	rout.HandleFunc("/api/requests", contr.Get).Methods("GET")
	rout.HandleFunc("/api/repeat/{id}", contr.Repeat).Methods("GET")
	rout.HandleFunc("/api/scan/{id}", contr.Scan).Methods("GET")

	return rout
}
