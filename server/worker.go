package server

import (
	"context"
	"runtime"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/labbcb/rnnr/proto"
	"github.com/pbnjay/memory"
	log "github.com/sirupsen/logrus"
)

// Worker struct wraps service info and Docker connection.
type Worker struct {
	proto.UnimplementedWorkerServer
	Info   *proto.Info
	Docker *Docker
}

// NewWorker creates a Worker.
// If cpuCores or ramGb is not defined (equal to 0) it will guess the available resources.
// It will warn if the defined values are bigger than guessed values.
func NewWorker(cpuCores int32, ramGb float64, volumes []string) (*Worker, error) {
	conn, err := DockerConnect(volumes)
	if err != nil {
		return nil, err
	}

	identifiedCpuCores := int32(runtime.NumCPU())
	if cpuCores == 0 {
		cpuCores = identifiedCpuCores
	}

	identifiedRamGb := float64(memory.TotalMemory() / 1e+9)
	if ramGb == 0 {
		ramGb = identifiedRamGb
	}

	worker := &Worker{
		Docker: conn,
		Info: &proto.Info{
			CpuCores:           cpuCores,
			RamGb:              ramGb,
			IdentifiedCpuCores: identifiedCpuCores,
			IdentifiedRamGb:    identifiedRamGb,
		},
	}

	return worker, nil
}

// GetInfo returns service info.
func (w *Worker) GetInfo(context.Context, *empty.Empty) (*proto.Info, error) {
	return w.Info, nil
}

// RunContainer starts a Docker container.
func (w *Worker) RunContainer(ctx context.Context, container *proto.Container) (*empty.Empty, error) {
	if err := w.Docker.Run(ctx, container); err != nil {
		log.WithError(err).WithFields(log.Fields{"id": container.Id, "image": container.Image}).Error("Unable to run container.")
		return nil, err
	}

	log.WithFields(log.Fields{"id": container.Id, "image": container.Image}).Info("Running container.")
	return &empty.Empty{}, nil
}

// CheckContainer checks if container is running.
func (w *Worker) CheckContainer(ctx context.Context, container *proto.Container) (*proto.State, error) {
	state, err := w.Docker.Check(ctx, container)
	if err != nil {
		log.WithError(err).WithField("id", container.Id).Error("Unable to check container.")
		return nil, err
	}

	if state.Exited {
		log.WithFields(log.Fields{"id": container.Id, "exitCode": state.ExitCode}).Info("Container exited.")
		w.Docker.RemoveContainer(ctx, container.Id)
	}

	return state, nil
}

// StopContainer stops and removes container.
func (w *Worker) StopContainer(ctx context.Context, container *proto.Container) (*empty.Empty, error) {
	if err := w.Docker.Stop(ctx, container.Id); err != nil {
		log.WithError(err).WithField("id", container.Id).Error("Unable to stop container.")
		return nil, err
	}

	log.WithField("id", container.Id).Info("Container stopped.")
	w.Docker.RemoveContainer(ctx, container.Id)
	return &empty.Empty{}, nil
}
