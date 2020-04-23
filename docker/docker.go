package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/labbcb/rnnr/pb"
	log "github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Docker struct wraps Docker client
type Docker struct {
	client *client.Client
}

// Connect creates a Docker client using environment variables
func Connect() (*Docker, error) {
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Docker{c}, nil
}

// Run runs a container
func (d *Docker) Run(ctx context.Context, container *pb.Container) error {
	if err := d.pullImage(container.Image, ioutil.Discard); err != nil {
		log.WithError(err).WithField("image", container.Image).Warn("Unable to pull image.")
	}

	var env []string
	for k, v := range container.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return d.runContainer(
		ctx,
		container.Id,
		container.Image,
		container.Command,
		container.WorkDir,
		env, mounts(container))
}

// Stop stops a container
func (d *Docker) Stop(ctx context.Context, id string) error {
	if err := d.client.ContainerStop(ctx, id, nil); err != nil {
		return err
	}

	return nil
}

// Check verifies if container is still running.
func (d *Docker) Check(ctx context.Context, container *pb.Container) (*pb.State, error) {
	resp, err := d.client.ContainerInspect(ctx, container.Id)
	if err != nil {
		return nil, err
	}

	if resp.State.Running {
		cpuTime, maxMem := d.getUsage(ctx, container.Id)
		return &pb.State{
			CpuTime: cpuTime,
			MaxMem:  maxMem,
		}, nil
	}

	stdout, stderr := d.getLogs(ctx, container.Id)
	return &pb.State{
		Exited:   true,
		ExitCode: int32(resp.State.ExitCode),
		Start:    asTimestamp(resp.State.StartedAt),
		End:      asTimestamp(resp.State.FinishedAt),
		Stdout:   stdout,
		Stderr:   stderr,
	}, nil
}

func (d *Docker) getLogs(ctx context.Context, id string) (stdout, stderr string) {
	var bout, berr bytes.Buffer

	out, err := d.client.ContainerLogs(ctx, id, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.WithError(err).WithField("id", id).Error("Unable to get logs.")
		return
	}

	_, err = stdcopy.StdCopy(&bout, &berr, out)
	if err != nil {
		log.WithError(err).WithField("id", id).Warn("Unable to demultiplex logs.")
		return
	}

	return bout.String(), berr.String()
}

// RemoveContainer removes a container.
func (d *Docker) RemoveContainer(ctx context.Context, id string) {
	if err := d.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true}); err != nil {
		log.WithError(err).Warn("Unable to remove container.")
	} else {
		log.WithField("id", id).Info("Container removed.")
	}
}

func asTimestamp(s string) *timestamp.Timestamp {
	t, _ := time.Parse(time.RFC3339Nano, s)
	p, _ := ptypes.TimestampProto(t)
	return p
}

func (d *Docker) pullImage(image string, w io.Writer) error {
	reader, err := d.client.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	_, err = io.Copy(w, reader)
	if err != nil {
		return err
	}
	return nil
}

func (d *Docker) runContainer(ctx context.Context, id, image string, command []string, workDir string, env []string, mounts []mount.Mount) error {
	resp, err := d.client.ContainerCreate(ctx, &container.Config{
		Image:      image,
		Cmd:        command,
		WorkingDir: workDir,
		Env:        env,
	}, &container.HostConfig{
		Mounts: mounts,
	}, nil, id)
	if err != nil {
		return err
	}

	return d.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
}

func (d *Docker) getUsage(ctx context.Context, id string) (cpuTime uint64, maxMem uint64) {
	resp, err := d.client.ContainerStats(ctx, id, false)
	if err != nil {
		log.WithError(err).WithField("id", id).Warn("Unable to get container stats.")
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	var stats types.Stats
	if json.NewDecoder(resp.Body).Decode(&stats) != nil {
		log.WithError(err).Warn("Unable to decode container stats.")
		return
	}

	return stats.CPUStats.CPUUsage.TotalUsage, stats.MemoryStats.MaxUsage
}

func mounts(t *pb.Container) []mount.Mount {
	var volumes []mount.Mount

	for _, input := range t.Outputs {
		volumes = addVolume(volumes, mount.Mount{
			Type:   mount.TypeBind,
			Source: filepath.Dir(input.HostPath),
			Target: filepath.Dir(input.ContainerPath),
		})
	}

	for _, input := range t.Inputs {
		volumes = addVolume(volumes, mount.Mount{
			Type:     mount.TypeBind,
			Source:   filepath.Dir(input.HostPath),
			Target:   filepath.Dir(input.ContainerPath),
			ReadOnly: true,
		})
	}

	return volumes
}

func addVolume(volumes []mount.Mount, v mount.Mount) []mount.Mount {
	// iterate over already added volumes to check if they are the same
	for i := range volumes {
		if volumes[i].Target == v.Target {
			return volumes
		}
		if strings.HasPrefix(volumes[i].Target, v.Target) {
			volumes[i] = v
			return volumes
		}
		if strings.HasPrefix(v.Target, volumes[i].Target) {
			return volumes
		}
	}

	return append(volumes, v)
}
