package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
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
)

// Docker struct wraps Docker client
type Docker struct {
	client *client.Client
}

// DockerConnect creates a Docker client using environment variables
func DockerConnect() (*Docker, error) {
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

	volumes := mounts(container)

	return d.runContainer(ctx, container.Id, container.Image, container.Command, container.WorkDir,
		env, volumes)
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

	state := pb.State{}
	if resp.State.Running {
		state.CpuPercent, state.CpuTime, state.Memory = d.getUsage(ctx, container.Id)
	} else {
		state.Exited = true
		state.ExitCode = int32(resp.State.ExitCode)
		state.Start = asTimestamp(resp.State.StartedAt)
		state.End = asTimestamp(resp.State.FinishedAt)
	}

	return &state, nil
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

func (d *Docker) getUsage(ctx context.Context, id string) (cpuPercent float64, cpuTime, memory uint64) {
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

	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	cpuPercent = (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	if math.IsNaN(cpuPercent) {
		cpuPercent = 0
	}

	return cpuPercent, stats.CPUStats.CPUUsage.TotalUsage, stats.MemoryStats.Stats["rss"]
}

func mounts(t *pb.Container) []mount.Mount {
	var volumes []mount.Mount

	for _, output := range t.Outputs {
		volumes = addVolume(volumes, output.HostPath, output.ContainerPath, false)
	}

	for _, input := range t.Inputs {
		volumes = addVolume(volumes, input.HostPath, input.ContainerPath, true)
	}

	return volumes
}

func addVolume(volumes []mount.Mount, hostPath, containerPath string, readOnly bool) []mount.Mount {

	hostDir := filepath.Dir(hostPath)
	containerDir := filepath.Dir(containerPath)

	for i := range volumes {
		// if container path is subpath of already added volume
		if strings.HasPrefix(containerPath, volumes[i].Target) {

			// if is read-only and host path is different of added volume
			if readOnly && hostDir != volumes[i].Source {
				return append(volumes, mount.Mount{
					Type:     mount.TypeBind,
					Source:   hostPath,
					Target:   containerPath,
					ReadOnly: readOnly,
				})
			}
			return volumes
		}

		// if volume is subpath of container path
		if strings.HasPrefix(volumes[i].Target, containerDir) {
			volumes[i].Source = hostDir
			volumes[i].Target = containerDir
			return volumes
		}
	}

	return append(volumes, mount.Mount{
		Type:     mount.TypeBind,
		Source:   hostDir,
		Target:   containerDir,
		ReadOnly: readOnly,
	})
}
