package worker

import (
	"context"
	"runtime"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/labbcb/rnnr/docker"
	"github.com/labbcb/rnnr/pb"
	"github.com/pbnjay/memory"
	log "github.com/sirupsen/logrus"
)

// Worker struct wraps service info and Docker connection.
type Worker struct {
	Info   *pb.Info
	Docker *docker.Docker
}

// New creates a Worker.
func New(cpuCores int32, ramGb float64, temp string) (*Worker, error) {
	conn, err := docker.Connect()
	if err != nil {
		return nil, err
	}

	if cpuCores == 0 {
		cpuCores = int32(runtime.NumCPU())
	}
	if ramGb == 0 {
		ramGb = float64(memory.TotalMemory() / 1e+9)
	}

	worker := &Worker{
		Docker: conn,
		Info: &pb.Info{
			CpuCores: cpuCores,
			RamGb:    ramGb,
		},
	}

	worker.Docker.Temp = temp

	return worker, nil
}

// GetInfo returns service info.
func (w *Worker) GetInfo(context.Context, *empty.Empty) (*pb.Info, error) {
	return w.Info, nil
}

// RunContainer starts a Docker container.
func (w *Worker) RunContainer(ctx context.Context, container *pb.Container) (*empty.Empty, error) {
	if err := w.Docker.Run(ctx, container); err != nil {
		log.WithError(err).WithFields(log.Fields{"id": container.Id, "image": container.Image}).Error("Unable to run container.")
		return nil, err
	}

	log.WithFields(log.Fields{"id": container.Id, "image": container.Image}).Info("Running container.")
	return &empty.Empty{}, nil
}

// CheckContainer checks if container is running.
func (w *Worker) CheckContainer(ctx context.Context, container *pb.Container) (*pb.State, error) {
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
func (w *Worker) StopContainer(ctx context.Context, container *pb.Container) (*empty.Empty, error) {
	if err := w.Docker.Stop(ctx, container.Id); err != nil {
		log.WithError(err).WithField("id", container.Id).Error("Unable to stop container.")
		return nil, err
	}

	log.WithField("id", container.Id).Info("Container stopped.")
	w.Docker.RemoveContainer(ctx, container.Id)
	return &empty.Empty{}, nil
}

func asTime(p *timestamp.Timestamp) time.Time {
	t, _ := ptypes.Timestamp(p)
	return t
}
