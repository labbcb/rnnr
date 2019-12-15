package worker

import (
	"fmt"
	"github.com/labbcb/rnnr/models"
	"log"
	"runtime"
	"time"

	"github.com/labbcb/rnnr/db"

	"github.com/labbcb/rnnr/docker"
	"github.com/labbcb/rnnr/server"
	"github.com/pbnjay/memory"
)

// Worker server is a standalone task executor that can be connected with a Master server.
type Worker struct {
	*server.Server
	Info *models.Info
}

// New creates a standalone worker server and initializes TES API endpoints.
// If cpuCores of ramGb is zero then the function will guess the maximum values.
func New(uri string, cpuCores int, ramGb float64) (*Worker, error) {
	client, err := db.Connect(uri, "rnnr-worker")
	if err != nil {
		return nil, fmt.Errorf("unable to connect to MongoDB: %w", err)
	}

	rnnr, err := docker.Connect()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Docker: %w", err)
	}

	if cpuCores == 0 {
		cpuCores = runtime.NumCPU()
	}
	if ramGb == 0 {
		ramGb = float64(memory.TotalMemory() / 1e+9)
	}

	worker := &Worker{
		Server: server.New(client, rnnr),
		Info: &models.Info{
			CPUCores: cpuCores,
			RAMGb:    ramGb,
		},
	}

	worker.register()
	go worker.StartMonitor(5 * time.Second)
	return worker, nil
}

func (w *Worker) StartMonitor(sleepTime time.Duration) {
	for {
		if err := w.InitializeAndRunTasks(); err != nil {
			log.Println(err)
		}

		if err := w.CheckTasks(); err != nil {
			log.Println(err)
		}

		time.Sleep(sleepTime)
	}
}

// InitializeAndRunTasks will iterate over queued tasks initializing and executing them.
func (w *Worker) InitializeAndRunTasks() error {
	ts, err := w.DB.FindByState(models.Queued)
	if err != nil {
		return fmt.Errorf("could not get queued tasks: %w", err)
	}

	for _, t := range ts {
		t.State = models.Initializing
		if err := w.DB.Update(t); err != nil {
			log.Printf("could not update state of task %s: %v", t.ID, err)
		}

		log.Println(t)
		go w.RunTask(t)
	}

	return nil
}

func (w *Worker) CheckTasks() error {
	runningTasks, err := w.DB.FindByState(models.Running)
	if err != nil {
		return fmt.Errorf("getting running tasks: %w", err)
	}

	for _, t := range runningTasks {
		go w.CheckTask(t)
	}

	return nil
}

func (w *Worker) RunTask(t *models.Task) {
	if err := w.Runner.Run(t); err != nil {
		log.Println(err)
		t.State = models.ExecutorError
		t.Logs = &models.Log{}
		t.Logs.EndTime = time.Now()
		t.Logs.SystemLogs = append(t.Logs.SystemLogs, err.Error())
	}
	if err := w.DB.Update(t); err != nil {
		log.Println("unable to update task:", err)
	}
	log.Println(t)
}

func (w *Worker) CheckTask(t *models.Task) {
	if err := w.Runner.Check(t); err != nil {
		log.Println(err)
		t.State = models.ExecutorError
		t.Logs = &models.Log{}
		t.Logs.EndTime = time.Now()
		t.Logs.SystemLogs = append(t.Logs.SystemLogs, err.Error())
	}
	if err := w.DB.Update(t); err != nil {
		log.Println("unable to update task:", err)
	}
	if t.State != models.Running {
		log.Println(t)
	}
}
