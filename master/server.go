package master

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/labbcb/rnnr/models"
	log "github.com/sirupsen/logrus"
	"time"
)

func (m *Master) CreateTask(t *models.Task) error {
	switch len(t.Executors) {
	case 0:
		return errors.New("no executors submitted")
	case 1:
	default:
		return fmt.Errorf("multiple executors not supported, got %d executors", len(t.Executors))
	}

	t.ID = uuid.New().String()
	t.CreationTime = time.Now()
	t.State = models.Queued
	t.Logs = &models.Log{}
	if t.Resources.CPUCores == 0 {
		t.Resources.CPUCores = 1
	}
	if err := m.DB.SaveTask(t); err != nil {
		return fmt.Errorf("unable to save models %m: %w", t.ID, err)
	}

	return nil
}

func (m *Master) GetTask(id string) (*models.Task, error) {
	t, err := m.DB.GetTask(id)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (m *Master) CancelTask(id string) error {
	task, err := m.GetTask(id)
	if err != nil {
		return err
	}

	if !task.Active() {
		return nil
	}

	if task.State == models.Queued || task.State == models.Initializing {
		task.State = models.Canceled
		task.Logs.EndTime = time.Now()
		return m.DB.UpdateTask(task)
	}

	node, err := m.DB.GetNode(task.RemoteHost)
	if err != nil {
		return err
	}
	if err := RemoteCancel(task, node); err != nil {
		log.WithField("error", err).Error("Unable to cancel task remotely.")
	}

	return m.DB.UpdateTask(task)
}

func (m *Master) ListTasks(namePrefix string, pageSize int, pageToken string, view models.View) (*models.ListTasksResponse, error) {
	ts, err := m.DB.AllTasks()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve all tasks: %w", err)
	}
	return &models.ListTasksResponse{Tasks: ts}, nil
}
