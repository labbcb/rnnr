package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/labbcb/rnnr/models"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

// Register binds endpoints for node management
func (m *Main) register() {
	m.Router.HandleFunc("/v1/nodes", m.handleListNodes()).Methods(http.MethodGet)
	m.Router.HandleFunc("/v1/nodes", m.handleEnableNode()).Methods(http.MethodPost)
	m.Router.HandleFunc("/v1/nodes/{id}", m.handleGetNode()).Methods(http.MethodGet)
	m.Router.HandleFunc("/v1/nodes/{id}:disable", m.handleDisableNode()).Methods(http.MethodPost)

	m.Router.HandleFunc("/ga4gh/tes/v1/tasks", m.handleListTasks()).Methods(http.MethodGet)
	m.Router.HandleFunc("/ga4gh/tes/v1/tasks", m.handleCreateTask()).Methods(http.MethodPost)
	m.Router.HandleFunc("/ga4gh/tes/v1/tasks/{id}", m.handleGetTask()).Methods(http.MethodGet)
	m.Router.HandleFunc("/ga4gh/tes/v1/tasks/{id}:cancel", m.handleCancelTask()).Methods(http.MethodPost)
	m.Router.HandleFunc("/ga4gh/tes/v1/service-info", m.handleGetServiceInfo()).Methods(http.MethodGet)
}

func (m *Main) handleListNodes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query()
		active, _ := strconv.ParseBool(v.Get("active"))
		var activeP *bool
		if active {
			activeP = &active
		}

		nodes, err := m.ListNodes(activeP)
		if err != nil {
			log.WithField("error", err).Error("Unable to get all nodes.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := m.UpdateNodesWorkload(nodes); err != nil {
			log.WithField("error", err).Error("Unable to update nodes workload.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		encodeJSON(w, nodes)
	}
}

func (m *Main) handleEnableNode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var node models.Node
		if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
			log.WithField("error", err).Error("Unable to decode JSON.")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := m.EnableNode(&node); err != nil {
			log.WithFields(log.Fields{"host": node.Host, "port": node.Port, "error": err}).Error("Unable to enable node.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.WithFields(log.Fields{"host": node.Host, "port": node.Port, "cpu": node.CPUCores, "ram": node.RAMGb}).Info("Node enabled.")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := map[string]string{"host": node.Host}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.WithField("error", err).Error("Unable to encode JSON.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (m *Main) handleGetNode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		node, err := m.DB.GetNode(id)
		if err != nil {
			log.WithFields(log.Fields{"id": id, "error": err}).Error("Unable to get node.")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if err := m.UpdateNodesWorkload([]*models.Node{node}); err != nil {
			log.WithField("error", err).Error("Unable to update nodes workload.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		encodeJSON(w, node)
	}
}

func (m *Main) handleDisableNode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cancel bool
		if err := json.NewDecoder(r.Body).Decode(&cancel); err != nil {
			log.WithField("error", err).Error("Unable to decode JSON.")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		host := mux.Vars(r)["id"]
		if err := m.DisableNode(host, cancel); err != nil {
			log.WithFields(log.Fields{"host": host, "error": err}).Error("Unable to disable node.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.WithFields(log.Fields{"host": host}).Info("Node disabled.")
	}
}

func (m *Main) handleCreateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var task models.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			log.WithField("error", err).Error("Unable to decode JSON.")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := m.CreateTask(&task); err != nil {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to create task.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name}).Info("Task created.")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(&models.CreateTaskResponse{ID: task.ID}); err != nil {
			log.WithField("error", err).Error("Unable to encode JSON.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (m *Main) handleGetTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		v := r.URL.Query()

		viewString := v.Get("view")
		if viewString == "" {
			viewString = "MINIMAL"
		}
		view := models.View(strings.ToUpper(viewString))

		task, err := m.GetTask(id, view)
		if err != nil {
			log.WithFields(log.Fields{"id": id, "error": err}).Error("Unable to get task.")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		encodeJSON(w, task)
	}
}

func (m *Main) handleListTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query()
		namePrefix := v.Get("name_prefix")
		pageSize, _ := strconv.ParseInt(v.Get("page_size"), 10, 64)
		pageToken, _ := strconv.ParseInt(v.Get("page_token"), 10, 64)

		viewString := v.Get("view")
		if viewString == "" {
			viewString = "MINIMAL"
		}
		view := models.View(strings.ToUpper(viewString))

		var states []models.State
		for _, state := range v["state"] {
			states = append(states, models.State(state))
		}

		nodes := v["node"]

		tasks, err := m.ListTasks(namePrefix, pageSize, pageToken, view, nodes, states)
		if err != nil {
			log.WithFields(log.Fields{"pageSize": pageSize, "pageToken": pageToken, "view": view, "error": err}).Fatal("Unable to get all tasks.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		encodeJSON(w, tasks)
	}
}

func (m *Main) handleCancelTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if err := m.CancelTask(id); err != nil {
			log.WithFields(log.Fields{"id": id, "error": err}).Error("Unable to cancel task.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.WithField("id", id).Info("Task canceled.")
		encodeJSON(w, new(models.CancelTaskResponse))
	}
}

func (m *Main) handleGetServiceInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encodeJSON(w, m.ServiceInfo)
	}
}

func encodeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.WithFields(log.Fields{"error": err, "object": v}).Error("Unable to encode JSON.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
