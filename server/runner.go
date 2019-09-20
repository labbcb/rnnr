package server

import "github.com/labbcb/rnnr/task"

type Runner interface {
	Run(*task.Task) error
	Check(*task.Task) error
	Cancel(*task.Task) error
}
