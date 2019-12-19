package master

import (
	"fmt"
	"time"

	"github.com/labbcb/rnnr/client"
	"github.com/labbcb/rnnr/models"
)

// Remote executes tasks in computing nodes remotely.
type Remote struct {
}

// Run submits a task to a remote computing node.
func (r *Remote) Run(t *models.Task) error {
	id, err := client.CreateTask(t.RemoteHost, t)
	if err != nil {
		if _, ok := err.(*client.NetworkError); ok {
			t.State = models.Queued
			t.RemoteHost = ""
		} else {
			t.State = models.ExecutorError
			t.Logs = &models.Log{}
			t.Logs.EndTime = time.Now()
			t.Logs.SystemLogs = append(t.Logs.SystemLogs, err.Error())
		}
		return fmt.Errorf("unable to run models %s at %s: %w", t.ID, t.RemoteHost, err)
	}

	t.RemoteTaskID = id
	t.State = models.Running
	return nil
}

// Check requests remote task and updates local task.
func (r *Remote) Check(t *models.Task) error {
	remoteTask, err := client.GetTask(t.RemoteHost, t.RemoteTaskID)
	if err != nil {
		if _, ok := err.(*client.NetworkError); ok {
			t.State = models.Queued
			t.RemoteHost = ""
		} else {
			t.State = models.ExecutorError
			t.Logs = &models.Log{}
			t.Logs.EndTime = time.Now()
			t.Logs.SystemLogs = append(t.Logs.SystemLogs, err.Error())
		}
		return fmt.Errorf("unable to check models %s: %w", t.ID, err)
	}

	if remoteTask.Active() {
		return nil
	}

	t.State = remoteTask.State
	t.Executors = remoteTask.Executors
	t.Logs = remoteTask.Logs
	t.Outputs = remoteTask.Outputs
	return nil
}

// Cancel request task cancellation to node and keep requesting remote task until its state is changed.
func (r *Remote) Cancel(t *models.Task) error {
	if !t.Active() {
		return fmt.Errorf("models %s is not active, it is %s", t.ID, t.State)
	}

	if t.State == models.Queued || t.State == models.Initializing {
		t.State = models.Canceled
		t.Logs = &models.Log{}
		t.Logs.EndTime = time.Now()
		return nil
	}

	if err := client.CancelTask(t.RemoteHost, t.RemoteTaskID); err != nil {
		t.State = models.ExecutorError
		t.Logs = &models.Log{}
		t.Logs.EndTime = time.Now()
		t.Logs.SystemLogs = append(t.Logs.SystemLogs, err.Error())
		return fmt.Errorf("unable to cancel models %s at %s: %w", t.ID, t.RemoteHost, err)
	}

	remoteTask, err := client.GetTask(t.RemoteHost, t.RemoteTaskID)
	if err != nil {
		t.State = models.ExecutorError
		t.Logs = &models.Log{}
		t.Logs.EndTime = time.Now()
		t.Logs.SystemLogs = append(t.Logs.SystemLogs, err.Error())
		return fmt.Errorf("unable to get canceled models %s at %s: %w", t.ID, t.RemoteHost, err)
	}

	t.State = remoteTask.State
	t.Executors = remoteTask.Executors
	t.Logs = remoteTask.Logs
	t.Outputs = remoteTask.Outputs
	return nil
}
