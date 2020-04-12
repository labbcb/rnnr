package master

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/labbcb/rnnr/models"
	"github.com/labbcb/rnnr/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetNodeResources gets node resource information.
func GetNodeResources(node *models.Node) (int32, float64, error) {
	conn, err := grpc.Dial(node.Address(), grpc.WithInsecure())
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.WithError(err).Fatal("Unable to close client connection.")
		}
	}()

	info, err := pb.NewWorkerClient(conn).GetInfo(context.Background(), &empty.Empty{})
	if err != nil {
		return 0, 0, err
	}
	return info.CpuCores, info.RamGb, nil
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
	_, err = pb.NewWorkerClient(conn).RunContainer(context.Background(), asContainer(task))
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

	state, err := pb.NewWorkerClient(conn).CheckContainer(context.Background(), asContainer(task))
	if err != nil {
		if status.Code(err) == codes.Unavailable {
			return NetworkError{err}
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
		task.Logs.ExecutorLogs = executorLogs(state)
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

	_, err = pb.NewWorkerClient(conn).StopContainer(context.Background(), asContainer(task))
	if status.Code(err) == codes.Unavailable {
		return &NetworkError{err}
	}
	return err
}

func asContainer(t *models.Task) *pb.Container {
	return &pb.Container{
		Id:      t.ID,
		Image:   t.Executors[0].Image,
		Command: t.Executors[0].Command,
		WorkDir: t.Executors[0].WorkDir,
		Outputs: outputs(t.Outputs),
		Inputs:  inputs(t.Inputs),
		Env:     t.Executors[0].Env,
	}
}

func outputs(os []*models.Output) []*pb.Volume {
	var vs []*pb.Volume
	for _, o := range os {
		vs = append(vs, &pb.Volume{
			HostPath:      o.URL,
			ContainerPath: o.Path,
		})
	}

	return vs
}

func inputs(is []*models.Input) []*pb.Volume {
	var vs []*pb.Volume
	for _, i := range is {
		vs = append(vs, &pb.Volume{
			HostPath:      i.URL,
			ContainerPath: i.Path,
		})
	}

	return vs
}

func executorLogs(m *pb.State) []*models.ExecutorLog {
	return []*models.ExecutorLog{{
		StartTime: asTime(m.Start),
		EndTime:   asTime(m.End),
		Stdout:    m.Stdout,
		Stderr:    m.Stderr,
		ExitCode:  m.ExitCode,
	}}
}

func asTime(p *timestamp.Timestamp) time.Time {
	t, _ := ptypes.Timestamp(p)
	return t
}
