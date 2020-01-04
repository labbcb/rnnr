package master

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/labbcb/rnnr/db"
	"github.com/labbcb/rnnr/models"
	log "github.com/sirupsen/logrus"
	"time"
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

func (m *Master) StartMonitor(sleepTime time.Duration) {
	for {
		if err := m.RunTasks(); err != nil {
			log.Error(err)
		}

		if err := m.CheckTasks(); err != nil {
			log.Error(err)
		}

		time.Sleep(sleepTime)
	}
}

func (m *Master) RunTasks() error {
	cursor, err := m.DB.FindByState(models.Queued)
	if err != nil {
		return err
	}
	defer func() {
		if err := cursor.Close(nil); err != nil {
			log.Fatal(err)
		}
	}()

	var task *models.Task
	for cursor.Next(nil) {
		if err := cursor.Decode(&task); err != nil {
			log.WithField("error", err).Error("Unable to decode BSON.")
			continue
		}

		node, err := m.RequestNode(task.Resources)
		switch err.(type) {
		case nil:
			task.RemoteHost = node.Host
			task.State = models.Initializing
			if err := m.DB.UpdateTask(task); err != nil {
				log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to update task.")
				continue
			}
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost}).Info("Task initialized.")

			go m.RunTask(task, node)
		case NoActiveNodes:
		case NoEnoughResources:
		default:
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to request node.")
		}
	}

	return nil
}

// RunTask will check computing node and delegate task execution to the node.
// If node is not active or not responding them the task will be put as queued.
// Node will be disabled if not responding.
func (m *Master) RunTask(task *models.Task, node *models.Node) {
	if err := RemoteRun(task, node); err != nil {
		if !node.Active {
			if err := m.DisableNode(node.Host); err != nil {
				log.WithFields(log.Fields{"host": node.Host, "error": err}).Error("Unable to disable node.")
			}
			log.WithFields(log.Fields{"host": node.Host, "error": err}).Warn("Unreachable node. Disabled.")
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost}).Warn("Putting the task back in the queue.")
		} else {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost, "state": task.State, "error": err}).Error("Unable to run task.")
		}
		if err := m.DB.UpdateTask(task); err != nil {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to update task.")
		}
		return
	}

	if err := m.DB.UpdateTask(task); err != nil {
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to update task.")
	}

	log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost}).Info("Running task.")
}

// CheckTasks will iterate over running tasks checking if they have been completed well or not.
// It runs concurrently.
func (m *Master) CheckTasks() error {
	cursor, err := m.DB.FindByState(models.Running)
	if err != nil {
		return err
	}
	defer func() {
		if err := cursor.Close(nil); err != nil {
			log.Fatal(err)
		}
	}()

	var task models.Task
	for cursor.Next(nil) {
		if err := cursor.Decode(&task); err != nil {
			log.WithField("error", err).Error("Unable to decode BSON.")
			continue
		}

		m.CheckTask(&task)
	}

	return nil
}

// CheckTask will check is a given tasks has been completed.
func (m *Master) CheckTask(task *models.Task) {
	node, err := m.DB.GetNode(task.RemoteHost)
	if err != nil {
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost}).Error("Unable to get node.")
		m.enqueueTask(task)
		return
	}

	if !node.Active {
		m.enqueueTask(task)
		return
	}

	if err := RemoteCheck(task, node); err != nil {
		if !node.Active {
			if err := m.DisableNode(node.Host); err != nil {
				log.WithFields(log.Fields{"host": node.Host, "error": err}).Error("Unable to disable node.")
			}
			log.WithFields(log.Fields{"host": node.Host, "error": err}).Warn("Unreachable node. Disabled.")
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost}).Warn("Putting the task back in the queue.")
		} else {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost, "state": task.State, "error": err}).Error("Unable to check task.")
		}
		if err := m.DB.UpdateTask(task); err != nil {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to update task.")
		}
		return
	}

	if task.State != models.Running {
		elapsed := task.Logs.EndTime.Sub(task.Logs.StartTime)
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost, "state": task.State, "elapsed": elapsed}).Info("Task finished.")
	}

	if err := m.DB.UpdateTask(task); err != nil {
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to update task.")
	}
}

func (m *Master) enqueueTask(task *models.Task) {
	task.State = models.Queued
	task.RemoteHost = ""
	if err := m.DB.UpdateTask(task); err != nil {
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to update task.")
	}
}
