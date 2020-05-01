package master

import (
	"fmt"
	"time"

	"github.com/gorilla/mux"
	"github.com/labbcb/rnnr/db"
	"github.com/labbcb/rnnr/models"
	log "github.com/sirupsen/logrus"
)

// Master is a master instance.
type Master struct {
	Router      *mux.Router
	DB          *db.DB
	ServiceInfo *models.ServiceInfo
}

// New creates a server and initializes TES API and Node management endpoints.
// database is URI to MongoDB (without database name, which is 'rnnr-master')
func New(database string, sleepTime time.Duration) (*Master, error) {
	connection, err := db.Connect(database, "rnnr")
	if err != nil {
		return nil, fmt.Errorf("connecting to MongoDB: %w", err)
	}

	master := &Master{
		Router: mux.NewRouter(),
		DB:     connection,
		ServiceInfo: &models.ServiceInfo{
			Name:    "rnnr",
			Doc:     "Distributed Task Executor for Genomics Research. GA4GH TES API implementation.",
			Storage: []string{"NFS"},
		},
	}
	master.register()
	go master.StartTaskManager(sleepTime)
	return master, nil
}

// StartTaskManager starts task management.
func (m *Master) StartTaskManager(sleepTime time.Duration) {
	for {
		if err := m.InitializeTasks(); err != nil {
			log.WithError(err).Warn("Unable to initialize tasks.")
		}
		if err := m.RunTasks(); err != nil {
			log.WithError(err).Warn("Unable to run tasks.")
		}

		if err := m.CheckTasks(); err != nil {
			log.WithError(err).Warn("Unable to check tasks.")
		}

		time.Sleep(sleepTime)
	}
}

// InitializeTasks iterates over all Queued tasks requesting a computing node for each task.
// The selected node is assigned to perform the task. The task changes to the Initializing state.
// If no active node has enough computing resources to perform the task the same is kept in queue.
func (m *Master) InitializeTasks() error {
	tasks, err := m.DB.FindByState(0, 0, models.Full, models.Queued)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		node, err := m.RequestNode(task.Resources)
		switch err.(type) {
		case nil:
			task.Worker = &models.Worker{
				Host: node.Host,
			}
			task.State = models.Initializing
			now := time.Now()
			task.Logs.StartTime = &now
			if err := m.DB.UpdateTask(task); err != nil {
				log.WithError(err).WithFields(log.Fields{"id": task.ID, "name": task.Name}).Error("Unable to update task.")
				continue
			}
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.Worker.Host}).Info("Task initialized.")
		case *NoActiveNodes:
			log.Warn("No active nodes")
		case *NoEnoughResources:
		default:
			log.WithError(err).WithFields(log.Fields{"id": task.ID, "name": task.Name}).Error("Unable to request node.")
		}
	}
	return nil
}

// RunTasks tries to start initialized tasks.
func (m *Master) RunTasks() error {
	tasks, err := m.DB.FindByState(0, 0, models.Full, models.Initializing)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		node, err := m.DB.GetNode(task.Worker.Host)
		if err != nil {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.Worker.Host}).Error("Unable to get node.")
			continue
		}

		task.State = models.Running
		if err := m.DB.UpdateTask(task); err != nil {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to update task.")
			continue
		}

		go m.RunTask(task, node)
	}
	return nil
}

// RunTask remotely starts a task.
func (m *Master) RunTask(task *models.Task, node *models.Node) {
	switch err := RemoteRun(task, node.Address()).(type) {
	case nil:
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.Worker.Host}).Info("Task running.")
	case *NetworkError:
		log.WithError(err).WithFields(log.Fields{"id": task.ID, "host": task.Worker.Host}).Warn("Network error.")
		m.enqueueTask(task)
	default:
		task.State = models.SystemError
		now := time.Now()
		task.Logs.EndTime = &now
		task.Logs.SystemLogs = []string{err.Error()}
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.Worker.Host, "state": task.State, "error": err}).Error("Unable to run task.")

		if err := m.DB.UpdateTask(task); err != nil {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to update task.")
		}
	}
}

// CheckTasks will iterate over running tasks checking if they have been completed well or not.
// It runs concurrently.
func (m *Master) CheckTasks() error {
	tasks, err := m.DB.FindByState(0, 0, models.Full, models.Running)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		node, err := m.DB.GetNode(task.Worker.Host)
		if err != nil {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.Worker.Host}).Error("Unable to get node.")
			continue
		}

		switch err := RemoteCheck(task, node.Address()).(type) {
		case nil:
			if task.State != models.Running {
				now := time.Now()
				task.Logs.EndTime = &now
				log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.Worker.Host, "state": task.State}).Info("Task finished.")
			}
		case *NetworkError:
			log.WithError(err).WithFields(log.Fields{"id": task.ID, "host": task.Worker.Host}).Warn("Network error.")
		default:
			task.State = models.SystemError
			task.Logs.SystemLogs = append(task.Logs.SystemLogs, err.Error())
			log.WithError(err).WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.Worker.Host, "state": task.State}).Error("Unable to check task.")
		}

		if err := m.DB.UpdateTask(task); err != nil {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to update task.")
		}
	}
	return nil
}
