package worker

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/labbcb/rnnr/docker"
	"github.com/labbcb/rnnr/pb"
	"github.com/pbnjay/memory"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime"
)

type Worker struct {
	Info   *pb.Info
	Docker *docker.Docker
}

func New(cpuCores int32, ramGb float64) (*Worker, error) {
	rnnr, err := docker.Connect()
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
		Docker: rnnr,
		Info: &pb.Info{
			CpuCores: cpuCores,
			RamGb:    ramGb,
		},
	}

	return worker, nil
}

func (w *Worker) GetInfo(context.Context, *empty.Empty) (*pb.Info, error) {
	return w.Info, nil
}

func (w *Worker) RunContainer(ctx context.Context, container *pb.Container) (*empty.Empty, error) {
	if err := w.Docker.Run(ctx, container); err != nil {
		log.WithFields(log.Fields{"id": container.Id, "image": container.Image, "error": err}).Error("Unable to run container.")
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.WithFields(log.Fields{"id": container.Id, "image": container.Image}).Info("Running container.")
	return &empty.Empty{}, nil
}

func (w *Worker) CheckContainer(ctx context.Context, container *pb.Container) (*pb.State, error) {
	state, err := w.Docker.Check(ctx, container)
	if err != nil {
		log.WithFields(log.Fields{"id": container.Id, "error": err}).Error("Unable to check container.")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if state.Exited {
		log.WithFields(log.Fields{"id": container.Id, "exitCode": state.ExitCode}).Info("Container exited.")
	}

	return state, nil
}

func (w *Worker) StopContainer(ctx context.Context, container *pb.Container) (*empty.Empty, error) {
	if err := w.Docker.Stop(ctx, container.Id); err != nil {
		log.WithFields(log.Fields{"id": container.Id, "error": err}).Error("Unable to stop container.")
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.WithFields(log.Fields{"id": container.Id}).Info("Stopped container.")
	return &empty.Empty{}, nil
}
