package master

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/labbcb/rnnr/node"
)

// Register binds endpoints for node management
func (m *Master) register() {
	m.Server.Router.HandleFunc("/nodes", m.handleListNodes()).Methods(http.MethodGet)
	m.Server.Router.HandleFunc("/nodes", m.handleActivateNode()).Methods(http.MethodPost)
	m.Server.Router.HandleFunc("/nodes/{id}", m.handleDeactivateNode()).Methods(http.MethodDelete)
}

func (m *Master) handleListNodes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ns, err := m.GetAllNodes()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := m.UpdateNodesWorkload(ns); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// encode nodes to json
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(&ns); err != nil {
			log.Printf("encoding nodes: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (m *Master) handleActivateNode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var n node.Node
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			log.Println("decoding request body to json:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := m.Activate(&n); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := map[string]string{"id": n.ID}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("encoding node id: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (m *Master) handleDeactivateNode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := mux.Vars(r)["id"]
		if err := m.Deactivate(host); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
