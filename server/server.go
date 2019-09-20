package server

import (
	"fmt"
	"github.com/labbcb/rnnr/db"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/labbcb/rnnr/task"
)

// Server is the configurable RNNR server.
type Server struct {
	Router      *mux.Router
	DB          *db.DB
	Runner      Runner
	ServiceInfo *task.ServiceInfo
}

// New creates a server instance and register TES API handlers.
func New(db *db.DB, rnnr Runner) *Server {
	s := &Server{
		Router: mux.NewRouter(),
		DB:     db,
		Runner: rnnr,
	}
	s.register()
	return s
}

// Create generates a task ID and set state to queued.
// Task is stored in database.
func (s *Server) Create(t *task.Task) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("unable to generate task id: %w", err)
	}
	t.ID = id.String()
	t.CreationTime = time.Now()
	t.State = task.Queued
	if t.Resources.CPUCores == 0 {
		t.Resources.CPUCores = 1
	}
	if err := s.DB.Save(t); err != nil {
		return fmt.Errorf("unable to save task %s: %w", t.ID, err)
	}
	log.Println(t)
	return nil
}

// Get search for a tasks by its ID returning it if exist.
func (s *Server) Get(id string) (*task.Task, error) {
	t, err := s.DB.Get(id)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve task %s: %w", id, err)
	}
	return t, nil
}

// Cancel stops an active task and updates its state in database.
func (s *Server) Cancel(id string) error {
	t, err := s.Get(id)
	if err != nil {
		return err
	}
	if !t.Active() {
		return nil
	}
	if err := s.Runner.Cancel(t); err != nil {
		return fmt.Errorf("unable to cancel task %s: %w", id, err)
	}
	if err := s.DB.Update(t); err != nil {
		return fmt.Errorf("unable to save canceled task %s: %w", id, err)
	}
	log.Println(t)
	return nil
}

// List returns all tasks stored in database.
func (s *Server) List(namePrefix string, pageSize int, pageToken string, view task.View) (*task.ListTasksResponse, error) {
	ts, err := s.DB.FindAll()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve all tasks: %w", err)
	}
	return &task.ListTasksResponse{Tasks: ts}, nil
}
