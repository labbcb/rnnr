package worker

import (
	"encoding/json"
	"log"
	"net/http"
)

func (w *Worker) register() {
	w.Server.Router.HandleFunc("/info", w.handleGetInfo()).Methods(http.MethodGet)
}

func (w *Worker) handleGetInfo() http.HandlerFunc {
	return func(writer http.ResponseWriter, r *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(&w.Info); err != nil {
			log.Printf("encoding node info: %v\n", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
