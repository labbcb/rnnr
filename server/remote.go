package server

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/labbcb/rnnr/models"
	"github.com/labbcb/rnnr/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetNodeResources gets node resource information.
func GetNodeResources(node *models.Node) (*proto.Info, error) {
	conn, err := grpc.Dial(node.Address(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.WithError(err).Fatal("Unable to close client connection.")
		}
	}()

	info, err := proto.NewWorkerClient(conn).GetInfo(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}
	return info, nil
}

// RemoteRun remotely runs a task as a container.
func RemoteRun(task *models.Task, address string) error {
	// create a connection with worker node
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return &NetworkError{err}
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.WithError(err).Fatal("Unable to close client connection.")
		}
	}()

	// convert a task to a container and remotely runs it
	_, err = proto.NewWorkerClient(conn).RunContainer(context.Background(), asContainer(task))
	if status.Code(err) == codes.Unavailable {
		return &NetworkError{err}
	}
	return err
}

// RemoteCheck checks remotely a task.
func RemoteCheck(task *models.Task, address string) error {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return &NetworkError{err}
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	state, err := proto.NewWorkerClient(conn).CheckContainer(context.Background(), asContainer(task))
	if err != nil {
		if status.Code(err) == codes.Unavailable {
			return &NetworkError{err}
		}
		return err
	}

	// task finished
	if state.Exited {
		if state.ExitCode == 0 {
			task.State = models.Complete
		} else {
			task.State = models.ExecutorError
		}
		task.Logs[0].ExecutorLogs = executorLogs(state)
	} else {
		// update worker stats
		task.Metrics.CPUTime = state.CpuTime
		task.Metrics.CPUPercentage = state.CpuPercent
		if state.CpuTime > task.Metrics.CPUTime {
			task.Metrics.CPUTime = state.CpuTime
		}
		if state.Memory > task.Metrics.Memory {
			task.Metrics.Memory = state.Memory
		}
	}
	return nil
}

// RemoteCancel cancels remotely a task.
func RemoteCancel(task *models.Task, node *models.Node) error {
	conn, err := grpc.Dial(node.Address(), grpc.WithInsecure())
	if err != nil {
		return &NetworkError{err}
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	_, err = proto.NewWorkerClient(conn).StopContainer(context.Background(), asContainer(task))
	if status.Code(err) == codes.Unavailable {
		return &NetworkError{err}
	}
	return err
}

func asContainer(t *models.Task) *proto.Container {
	return &proto.Container{
		Id:      t.ID,
		Image:   t.Executors[0].Image,
		Command: t.Executors[0].Command,
		WorkDir: t.Executors[0].WorkDir,
		Outputs: outputs(t.Outputs),
		Inputs:  inputs(t.Inputs),
		Env:     t.Executors[0].Env,
	}
}

func outputs(os []*models.Output) []*proto.Volume {
	var vs []*proto.Volume
	for _, o := range os {
		vs = append(vs, &proto.Volume{
			HostPath:      o.URL,
			ContainerPath: o.Path,
		})
	}

	return vs
}

func inputs(is []*models.Input) []*proto.Volume {
	var vs []*proto.Volume
	for _, i := range is {
		vs = append(vs, &proto.Volume{
			HostPath:      i.URL,
			ContainerPath: i.Path,
		})
	}

	return vs
}

func executorLogs(state *proto.State) []*models.ExecutorLog {
	logs := &models.ExecutorLog{
		StartTime: state.Start.AsTime(),
		EndTime:   state.End.AsTime(),
		Stdout:    state.Stdout,
		Stderr:    state.Stderr,
		ExitCode:  state.ExitCode,
	}
	return []*models.ExecutorLog{logs}
}
