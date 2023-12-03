// Package container contains functions and methods for starting a container.
package container

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DatabaseOptions interface {
	ImageName() string
	ContainerName() string
	DataSourceName() string
	ContainerConfig() *container.Config
}

type ContainerExistsError struct {
	id   string
	name string
}

func NewContainerExistsError(name, id string) *ContainerExistsError {
	return &ContainerExistsError{
		id:   id,
		name: name,
	}
}

func (e *ContainerExistsError) Error() string {
	return fmt.Sprintf("container %q with ID %q exists", e.name, e.id)
}

func ContainerExists(ctx context.Context, cli *client.Client, containerName string) (string, error) {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return "", err
	}
	for _, container := range containers {
		for _, name := range container.Names {
			if strings.TrimLeft(name, "/") == containerName {
				return container.ID, NewContainerExistsError(name, container.ID)
			}
		}
	}
	return "", nil
}

func StartContainer(ctx context.Context, cli *client.Client, config *container.Config, imageName, containerName string) (string, error) {
	reader, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)

	containerID, err := ContainerExists(ctx, cli, containerName)
	var cee *ContainerExistsError
	if err == nil {
		resp, err := cli.ContainerCreate(ctx, config, nil, nil, nil, containerName)
		if err != nil {
			return "", err
		}
		containerID = resp.ID
	} else if err != nil && !errors.As(err, &cee) {
		return "", err
	}

	if err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return containerID, nil
}

func ContainerIP(ctx context.Context, cli *client.Client, containerID string) (string, error) {
	resp, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", err
	}
	return resp.NetworkSettings.IPAddress, nil
}
