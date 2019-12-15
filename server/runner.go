package server

import "github.com/labbcb/rnnr/models"

type Runner interface {
	Run(*models.Task) error
	Check(*models.Task) error
	Cancel(*models.Task) error
}
