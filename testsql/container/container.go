package container

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

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

func ContainerExists(ctx context.Context, t *testing.T, cli *client.Client, containerName string) (string, bool) {
	t.Helper()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, container := range containers {
		for _, name := range container.Names {
			if strings.TrimLeft(name, "/") == containerName {
				return container.ID, true
			}
		}
	}
	return "", false
}

func StartContainer(ctx context.Context, t *testing.T, cli *client.Client, config *container.Config, imageName, containerName string) string {
	t.Helper()
	reader, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)

	containerID, exists := ContainerExists(ctx, t, cli, containerName)
	if !exists {
		resp, err := cli.ContainerCreate(ctx, config, nil, nil, nil, containerName)
		if err != nil {
			t.Fatal(err)
		}
		containerID = resp.ID
	}

	if err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		t.Fatal(err)
	}

	return containerID
}

func ContainerIP(ctx context.Context, t *testing.T, cli *client.Client, containerID string) string {
	t.Helper()
	resp, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		t.Fatal(err)
	}
	return resp.NetworkSettings.IPAddress
}
