package worker

import (
	"log"
	"time"

	"github.com/labbcb/rnnr/task"
)

// Start starts two monitoring systems that will
// check state of the tasks communicating with database
// and task runner.
// System 1 will iterate over queued tasks initializing
// and running them. Second system will check running
// tasks to check they were not completed yet.
// Both systems will sleep when iteration is completed
// and start over again.
// Every action happens concurrently.
func (w *Worker) Start(sleepTime time.Duration) {
	// initialize and run queued tasks
	go func() {
		for {
			ts, err := w.DB.FindByState(task.Queued)
			if err != nil {
				log.Println("could not get queued tasks:", err)
				continue
			}
			for _, t := range ts {
				t.State = task.Initializing
				if err := w.DB.Update(t); err != nil {
					log.Printf("could not update state of task %s: %v", t.ID, err)
				}
				log.Println(t)
				go w.run(t)
			}
			time.Sleep(sleepTime)
		}
	}()
	// check running tasks
	go func() {
		for {
			ts, err := w.DB.FindByState(task.Running)
			if err != nil {
				log.Println("getting running tasks:", err)
				continue
			}
			for _, t := range ts {
				go w.check(t)
			}
			time.Sleep(sleepTime)
		}
	}()
}

func (w *Worker) run(t *task.Task) {
	if err := w.Runner.Run(t); err != nil {
		log.Println(err)
		t.State = task.ExecutorError
		t.Logs = &task.Log{}
		t.Logs.EndTime = time.Now()
		t.Logs.SystemLogs = append(t.Logs.SystemLogs, err.Error())
	}
	if err := w.DB.Update(t); err != nil {
		log.Println("unable to update task:", err)
	}
	log.Println(t)
}

func (w *Worker) check(t *task.Task) {
	if err := w.Runner.Check(t); err != nil {
		log.Println(err)
		t.State = task.ExecutorError
		t.Logs = &task.Log{}
		t.Logs.EndTime = time.Now()
		t.Logs.SystemLogs = append(t.Logs.SystemLogs, err.Error())
	}
	if err := w.DB.Update(t); err != nil {
		log.Println("unable to update task:", err)
	}
	if t.State != task.Running {
		log.Println(t)
	}
}
