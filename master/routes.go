package master

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labbcb/rnnr/models"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

// Register binds endpoints for node management
func (m *Master) register() {
	m.Router.HandleFunc("/nodes", m.handleListNodes()).Methods(http.MethodGet)
	m.Router.HandleFunc("/nodes", m.handleEnableNode()).Methods(http.MethodPost)
	m.Router.HandleFunc("/nodes/{id}:disable", m.handleDisableNode()).Methods(http.MethodPost)

	m.Router.HandleFunc("/tasks", m.handleListTasks()).Methods(http.MethodGet)
	m.Router.HandleFunc("/tasks", m.handleCreateTask()).Methods(http.MethodPost)
	m.Router.HandleFunc("/tasks/{id}", m.handleGetTask()).Methods(http.MethodGet)
	m.Router.HandleFunc("/tasks/{id}:cancel", m.handleCancelTask()).Methods(http.MethodPost)
	m.Router.HandleFunc("/tasks/service-info", m.handleGetServiceInfo()).Methods(http.MethodGet)

}

func (m *Master) handleListNodes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodes, err := m.GetAllNodes()
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

func (m *Master) handleEnableNode() http.HandlerFunc {
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

func (m *Master) handleDisableNode() http.HandlerFunc {
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

func (m *Master) handleCreateTask() http.HandlerFunc {
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

func (m *Master) handleGetTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		task, err := m.GetTask(id)
		if err != nil {
			log.WithFields(log.Fields{"id": id, "error": err}).Error("Unable to get task.")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		encodeJSON(w, task)
	}
}

func (m *Master) handleListTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pageSize, _ := strconv.Atoi(vars["pageSize"])
		pageToken := vars["pageToken"]
		view := models.View(vars["view"])
		tasks, err := m.ListTasks(vars["namePrefix"], pageSize, pageToken, view)
		if err != nil {
			log.WithFields(log.Fields{"pageSize": pageSize, "pageToken": pageToken, "view": view, "error": err}).Fatal("Unable to get all tasks.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		encodeJSON(w, tasks)
	}
}

func (m *Master) handleCancelTask() http.HandlerFunc {
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

func (m *Master) handleGetServiceInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encodeJSON(w, m.ServiceInfo)
	}
}

func encodeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.WithField("error", err).Error("Unable to encode JSON.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
