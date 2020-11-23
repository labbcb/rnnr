// Package server implements RNNR main logic to manage tasks and worker nodes.
package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labbcb/rnnr/models"
	log "github.com/sirupsen/logrus"
)

// CreateTask creates a task with new ID and queue state.
func (m *Main) CreateTask(t *models.Task) error {
	switch len(t.Executors) {
	case 0:
		return errors.New("no executors submitted")
	case 1:
	default:
		return fmt.Errorf("multiple executors not supported, got %d executors", len(t.Executors))
	}

	t.ID = uuid.New().String()
	t.State = models.Queued
	t.Logs = []*models.Log{}
	if t.Resources.CPUCores == 0 {
		t.Resources.CPUCores = 1
	}
	return m.DB.SaveTask(t)
}

// GetTask returns a task by its ID.
func (m *Main) GetTask(id string) (*models.Task, error) {
	t, err := m.DB.GetTask(id)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// CancelTask cancels a task by its ID.
func (m *Main) CancelTask(id string) error {
	task, err := m.GetTask(id)
	if err != nil {
		return err
	}

	if !task.Active() {
		return nil
	}

	if task.State == models.Queued || task.State == models.Initializing {
		task.State = models.Canceled
		now := time.Now()
		task.Logs[0].EndTime = &now
		return m.DB.UpdateTask(task)
	}

	node, err := m.DB.GetNode(task.Host)
	if err != nil {
		return err
	}
	if err := RemoteCancel(task, node); err != nil {
		log.WithField("error", err).Error("Unable to cancel task remotely.")
	}

	return m.DB.UpdateTask(task)
}

// ListTasks returns all tasks.
func (m *Main) ListTasks(namePrefix string, limit int64, start int64, view models.View, nodes []string, states []models.State) (*models.ListTasksResponse, error) {
	ts, err := m.DB.ListTasks(limit, start, view, nodes, states)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve all tasks: %w", err)
	}
	return &models.ListTasksResponse{Tasks: ts}, nil
}
