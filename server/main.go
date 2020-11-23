package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/labbcb/rnnr/models"
	log "github.com/sirupsen/logrus"
)

// Main is a main instance.
type Main struct {
	Router      *mux.Router
	DB          *DB
	ServiceInfo *models.ServiceInfo
}

// NewMain creates a server and initializes Task and Node endpoints.
// database is URI to MongoDB (without database name, which is rnnr).
// sleepTimes defines the time in seconds that main will sleep after task management iteration.
func NewMain(database string, sleepTime time.Duration) (*Main, error) {
	connection, err := MongoConnect(database, "rnnr")
	if err != nil {
		return nil, fmt.Errorf("connecting to MongoDB: %w", err)
	}

	main := &Main{
		Router: mux.NewRouter(),
		DB:     connection,
		ServiceInfo: &models.ServiceInfo{
			Name:    "rnnr",
			Doc:     "Distributed Task Executor for Genomics Research (https://bcblab.org/rnnr).",
			Storage: []string{"NFS"},
		},
	}
	main.register()
	go main.StartTaskManager(sleepTime)
	return main, nil
}

// StartTaskManager starts task management.
// It will iterate over: 1) queued tasks; 2) initialized tasks; and 3) running tasks.
// Then it will sleepTime seconds and start over.
func (m *Main) StartTaskManager(sleepTime time.Duration) {
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
func (m *Main) InitializeTasks() error {
	tasks, err := m.DB.ListTasks(0, 0, models.Full, nil, []models.State{models.Queued})
	if err != nil {
		return err
	}

	for _, task := range tasks {
		node, err := m.RequestNode(task.Resources)
		switch err.(type) {
		case nil:
			task.Host = node.Host
			task.State = models.Initializing
			now := time.Now()
			task.Logs[0].StartTime = &now
			if err := m.DB.UpdateTask(task); err != nil {
				log.WithError(err).WithFields(log.Fields{"id": task.ID, "name": task.Name}).Error("Unable to update task.")
				continue
			}
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.Host}).Info("Task initialized.")
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
func (m *Main) RunTasks() error {
	tasks, err := m.DB.ListTasks(0, 0, models.Full, nil, []models.State{models.Initializing})
	if err != nil {
		return err
	}

	ch := make(chan *models.Task)
	wg := &sync.WaitGroup{}
	wg.Add(len(tasks))
	for _, task := range tasks {
		node, err := m.DB.GetNode(task.Host)
		if err != nil {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.Host}).Error("Unable to get node.")
			continue
		}

		go m.RunTask(task, node, ch, wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for task := range ch {
		if err := m.DB.UpdateTask(task); err != nil {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Warn("Unable to update task.")
		}
	}

	return nil
}

// RunTask remotely starts a task.
func (m *Main) RunTask(task *models.Task, node *models.Node, res chan<- *models.Task, wg *sync.WaitGroup) {
	defer wg.Done()

	switch err := RemoteRun(task, node.Address()).(type) {
	case nil:
		task.State = models.Running
		task.Metrics = &models.Metrics{}
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.Host}).Info("Task running.")
	case *NetworkError:
		log.WithError(err).WithFields(log.Fields{"id": task.ID, "host": task.Host}).Warn("Network error.")
	default:
		task.State = models.SystemError
		now := time.Now()
		task.Logs[0].EndTime = &now
		task.Logs[0].SystemLogs = []string{err.Error()}
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.Host, "state": task.State, "error": err}).Error("Unable to run task.")
	}

	res <- task
}

// CheckTasks will iterate over running tasks checking if they have been completed well or not.
// It runs concurrently.
func (m *Main) CheckTasks() error {
	tasks, err := m.DB.ListTasks(0, 0, models.Full, nil, []models.State{models.Running})
	if err != nil {
		return err
	}

	ch := make(chan *models.Task)
	wg := &sync.WaitGroup{}
	wg.Add(len(tasks))
	for _, task := range tasks {
		node, err := m.DB.GetNode(task.Host)
		if err != nil {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.Host}).Error("Unable to get node.")
			continue
		}

		go CheckTask(task, node, ch, wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for task := range ch {
		if err := m.DB.UpdateTask(task); err != nil {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to update task.")
		}
	}

	return nil
}

// CheckTask remotely check a running task.
func CheckTask(task *models.Task, node *models.Node, res chan<- *models.Task, wg *sync.WaitGroup) {
	defer wg.Done()

	switch err := RemoteCheck(task, node.Address()).(type) {
	case nil:
		if task.State != models.Running {
			now := time.Now()
			task.Logs[0].EndTime = &now
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.Host, "state": task.State}).Info("Task finished.")
		}
	case *NetworkError:
		log.WithError(err).WithFields(log.Fields{"id": task.ID, "host": task.Host}).Warn("Network error.")
	default:
		task.State = models.SystemError
		task.Logs[0].SystemLogs = append(task.Logs[0].SystemLogs, err.Error())
		log.WithError(err).WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.Host, "state": task.State}).Error("Unable to check task.")
	}

	res <- task
}
