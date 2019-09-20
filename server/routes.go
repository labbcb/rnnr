package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/labbcb/rnnr/task"
)

func (s *Server) register() {
	s.Router.HandleFunc("/ga4gh/tes/v1/tasks", s.handleListTasks()).Methods(http.MethodGet)
	s.Router.HandleFunc("/ga4gh/tes/v1/tasks", s.handleCreateTask()).Methods(http.MethodPost)
	s.Router.HandleFunc("/ga4gh/tes/v1/tasks/{id}", s.handleGetTask()).Methods(http.MethodGet)
	s.Router.HandleFunc("/ga4gh/tes/v1/tasks/{id}:cancel", s.handleCancelTask()).Methods(http.MethodPost)
	s.Router.HandleFunc("/ga4gh/tes/v1/tasks/service-info", s.handleGetServiceInfo()).Methods(http.MethodGet)
}

func (s *Server) handleCreateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t task.Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			log.Println("decoding task from json:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := s.Create(&t); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(&task.CreateTaskResponse{ID: t.ID}); err != nil {
			log.Println("encoding response to json:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) handleGetTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get task by its ID
		id := mux.Vars(r)["id"]
		t, err := s.Get(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		encodeJSON(w, t)
	}
}

func (s *Server) handleListTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// call service to get all tasks
		vars := mux.Vars(r)
		pageSize, _ := strconv.Atoi(vars["pageSize"])
		ts, err := s.List(vars["namePrefix"], pageSize, vars["pageToken"], task.View(vars["view"]))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		encodeJSON(w, ts)
	}
}

func (s *Server) handleCancelTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if err := s.Cancel(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		encodeJSON(w, new(task.CancelTaskResponse))
	}
}

func (s *Server) handleGetServiceInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encodeJSON(w, s.ServiceInfo)
	}
}

func encodeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("encoding response to json:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
