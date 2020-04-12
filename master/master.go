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
func New(database string) (*Master, error) {
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
	go master.StartMonitor(5 * time.Second)
	return master, nil
}

// StartMonitor starts background tasks in loop.
func (m *Master) StartMonitor(sleepTime time.Duration) {
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

// InitializeTasks selects a worker node to execute queued tasks.
func (m *Master) InitializeTasks() error {
	tasks, err := m.DB.FindByState(models.Queued)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		node, err := m.RequestNode(task.Resources)
		switch err.(type) {
		case nil:
			task.RemoteHost = node.Host
			task.State = models.Initializing
			if err := m.DB.UpdateTask(task); err != nil {
				log.WithError(err).WithFields(log.Fields{"id": task.ID, "name": task.Name}).Warn("Unable to update task.")
				continue
			}
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost}).Info("Task initialized.")
		case NoActiveNodes:
		case NoEnoughResources:
		default:
			log.WithError(err).WithFields(log.Fields{"id": task.ID, "name": task.Name}).Error("Unable to request node.")
		}
	}
	return nil
}

// RunTasks tries to start initialized tasks.
func (m *Master) RunTasks() error {
	tasks, err := m.DB.FindByState(models.Initializing)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		go m.RunTask(task)
	}

	return nil
}

// RunTask will check computing node and delegate task execution to the node.
func (m *Master) RunTask(task *models.Task) {
	node, err := m.DB.GetNode(task.RemoteHost)
	if err != nil {
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost}).Error("Unable to get Enode.")
		return
	}

	task.Logs.StartTime = time.Now()
	switch err := RemoteRun(task, node.Address()).(type) {
	case nil:
		task.State = models.Running
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost}).Info("Task running.")
	case *NetworkError:
		log.WithError(err).WithFields(log.Fields{"id": task.ID, "host": task.RemoteHost}).Warn("Network error.")
	default:
		task.State = models.SystemError
		task.Logs.EndTime = time.Now()
		task.Logs.SystemLogs = []string{err.Error()}
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost, "state": task.State, "error": err}).Error("Unable to run task.")
	}

	if err := m.DB.UpdateTask(task); err != nil {
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to update task.")
	}
}

// CheckTasks will iterate over running tasks checking if they have been completed well or not.
// It runs concurrently.
func (m *Master) CheckTasks() error {
	tasks, err := m.DB.FindByState(models.Running)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		m.CheckTask(task)
	}

	return nil
}

// CheckTask will check is a given tasks has been completed.
func (m *Master) CheckTask(task *models.Task) {
	node, err := m.DB.GetNode(task.RemoteHost)
	if err != nil {
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost}).Error("Unable to get node.")
		return
	}

	switch err := RemoteCheck(task, node.Address()).(type) {
	case nil:
		if task.State != models.Running {
			task.Logs.EndTime = time.Now()
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost, "state": task.State}).Info("Task finished.")
		}
	case *NetworkError:
		log.WithError(err).WithFields(log.Fields{"id": task.ID, "host": task.RemoteHost}).Warn("Network error.")
	default:
		task.State = models.SystemError
		task.Logs.SystemLogs = append(task.Logs.SystemLogs, err.Error())
		log.WithError(err).WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost, "state": task.State}).Error("Unable to check task.")
	}

	if err := m.DB.UpdateTask(task); err != nil {
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to update task.")
	}
}
