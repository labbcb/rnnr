package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/labbcb/rnnr/models"
)

// Docker wraps Docker client to provide task-related operations
type Docker struct {
	client *client.Client
}

// Connect creates a Docker client
func Connect() (*Docker, error) {
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Docker{c}, nil
}

// Run creates a Docker container and execute it
func (d *Docker) Run(t *models.Task) error {
	// check how many executors have been submitted
	switch len(t.Executors) {
	case 0:
		return errors.New("no executors submitted")
	case 1:
	default:
		return fmt.Errorf("multiple executors not supported, got %d executors", len(t.Executors))
	}

	// wait until docker pulls image
	// since it does not return error if there are network issues,
	// it will try to create the container with local image
	if err := d.PullImage(t.Executors[0].Image, ioutil.Discard); err != nil {
		return fmt.Errorf("unable to pull image %s: %w", t.Executors[0].Image, err)
	}

	var env []string
	for k, v := range t.Executors[0].Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	cfg := &container.Config{
		Image:      t.Executors[0].Image,
		Cmd:        t.Executors[0].Command,
		WorkingDir: t.Executors[0].WorkDir,
		Env:        env,
	}

	// create Docker container
	resp, err := d.client.ContainerCreate(context.Background(), cfg, &container.HostConfig{
		Mounts: mounts(t),
	}, nil, t.ID)
	if err != nil {
		return err
	}

	// start container
	if err := d.client.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// update task
	t.State = models.Running
	t.Logs = &models.Log{
		ExecutorLogs: []*models.ExecutorLog{{StartTime: time.Now()}},
		StartTime:    time.Now(),
	}
	return nil
}

// Cancel stops and removes a running Docker container
func (d *Docker) Cancel(t *models.Task) error {
	// create context
	ctx := context.Background()
	if err := d.client.ContainerStop(ctx, t.ID, nil); err != nil {
		return err
	}
	if err := d.client.ContainerRemove(ctx, t.ID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}
	t.Logs.ExecutorLogs[0].EndTime = time.Now()
	// update models state
	t.Logs.EndTime = time.Now()
	t.State = models.Canceled
	return nil
}

// Check inspect a running Docker container if still running.
func (d *Docker) Check(t *models.Task) error {
	// do not update models when it is not running
	if t.State != models.Running {
		return nil
	}
	// create context
	ctx := context.Background()

	// inspect container
	json, err := d.client.ContainerInspect(ctx, t.ID)
	if err != nil {
		return err
	}
	// if container is not running anymore
	if !json.State.Running {
		startTime, err := time.Parse(time.RFC3339Nano, json.State.StartedAt)
		if err != nil {
			return err
		}
		endTime, err := time.Parse(time.RFC3339Nano, json.State.FinishedAt)
		if err != nil {
			return err
		}

		// check if executor terminated ok
		exitCode := json.State.ExitCode
		if exitCode != 0 {
			// signals that some of the executors did not terminated ok
			t.State = models.ExecutorError
		} else {
			t.State = models.Complete
		}

		// update executor log
		t.Logs.ExecutorLogs[0].StartTime = startTime
		t.Logs.ExecutorLogs[0].EndTime = endTime
		t.Logs.ExecutorLogs[0].ExitCode = exitCode
		t.Logs.EndTime = time.Now()
	}
	return nil
}

// PullImage tries to pull image from internet.
// If network is down it logs the error and returns no error.
func (d *Docker) PullImage(image string, w io.Writer) error {
	reader, err := d.client.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		log.Println(err)
		return nil
	}
	defer reader.Close()
	_, err = io.Copy(w, reader)
	if err != nil {
		return err
	}
	return nil
}

func mounts(t *models.Task) []mount.Mount {
	var volumes []mount.Mount

	for _, input := range t.Outputs {
		volumes = addVolume(volumes, mount.Mount{
			Type:   mount.TypeBind,
			Source: filepath.Dir(input.URL),
			Target: filepath.Dir(input.Path),
		})
	}

	for _, input := range t.Inputs {
		// when existing write content to file and skip bind mount
		if input.Content != "" {
			log.Printf("Intput.Content not supported. Input.Path: %s", input.Path)
			continue
		}
		volumes = addVolume(volumes, mount.Mount{
			Type:     mount.TypeBind,
			Source:   filepath.Dir(input.URL),
			Target:   filepath.Dir(input.Path),
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
