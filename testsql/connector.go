// Package testsql contains the primary mechanism for connecting to a database
// container for unit or other functional testing.
package testsql

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"

	tsqlcon "github.com/kevensen/go-testsql/testsql/container"

	"github.com/cenkalti/backoff"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/golang/glog"
)

type Database interface {
	DataSourceName(string) string
	ContainerImage() string
	ContainerName() string
	Port() int
	PortWithProtocol() string
	Env() []string
}

type ConnectionError error

type TestConnector struct {
	host   string
	dbConn Database
	cli    *client.Client
}

func New(ctx context.Context, t *testing.T, dbConn Database, host string) (*TestConnector, func()) {
	t.Helper()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatal(err)
	}

	testConn := &TestConnector{
		dbConn: dbConn,
		host:   host,
		cli:    cli,
	}

	cointainerID, cleanup, err := testConn.start(ctx, t)
	var connErr ConnectionError
	if err != nil && errors.As(err, &connErr) && host != "" {
		testConn.cli.ContainerRemove(ctx, cointainerID, types.ContainerRemoveOptions{Force: true})
		testConn.host = ""
		_, cleanup, err = testConn.start(ctx, t)
		if err != nil {
			t.Fatal()
		}

	} else if err != nil {
		t.Fatal(err)
	}

	return testConn, cleanup

}

func (tc *TestConnector) start(ctx context.Context, t *testing.T) (string, func(), error) {
	t.Helper()
	containerID, err := tsqlcon.StartContainer(ctx, tc.cli, tc.ContainerConfig(), tc.HostConfig(), tc.dbConn.ContainerImage(), tc.dbConn.ContainerName())
	if err != nil {
		return "", nil, err
	}
	if tc.host == "" {
		tc.host, err = tsqlcon.ContainerIP(ctx, tc.cli, containerID)
		if err != nil {
			return "", nil, err
		}
	}

	operation := func() error {
		options := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true}
		out, err := tc.cli.ContainerLogs(ctx, containerID, options)
		if err != nil {
			return fmt.Errorf("log error: %v", err)
		}
		buf := new(strings.Builder)
		_, err = io.Copy(buf, out)
		if err != nil {
			return fmt.Errorf("io copy error %v", err)
		}
		glog.V(3).InfoContext(ctx, buf.String())
		buf.Reset()
		return testConnection(ctx, tc.host, tc.dbConn.Port()) // or an error
	}

	err = backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		return "", nil, ConnectionError(err)
	}

	cleanup := func() {
		err := tc.cli.ContainerStop(ctx, containerID, container.StopOptions{})
		if err != nil {
			t.Fatal(err)
		}
		tc.cli.Close()
	}
	return containerID, cleanup, nil
}

func (tc *TestConnector) DataSourceName() string {
	return tc.dbConn.DataSourceName(tc.host)
}

func (tc *TestConnector) ContainerConfig() *container.Config {
	port := tc.dbConn.PortWithProtocol()
	return &container.Config{
		Image:        tc.dbConn.ContainerImage(),
		Tty:          false,
		ExposedPorts: nat.PortSet{nat.Port(port): struct{}{}},
		Env:          tc.dbConn.Env(),
	}
}

func (tc *TestConnector) HostConfig() *container.HostConfig {
	if tc.host != "" {
		return &container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port(tc.dbConn.PortWithProtocol()): []nat.PortBinding{
					{
						HostIP:   tc.host,
						HostPort: strconv.FormatInt(int64(tc.dbConn.Port()), 10),
					},
				},
			},
		}
	}
	return nil
}

func testConnection(ctx context.Context, host string, port int) error {
	endpoint := fmt.Sprintf("%s:%d", host, port)
	glog.V(1).InfoContext(ctx, "Attempting connection to: ", endpoint)
	conn, err := net.Dial("tcp", endpoint)
	if err != nil {
		glog.V(3).InfoContext(ctx, "Connection test failed to : %v", endpoint, ":", err)
		return fmt.Errorf("connection test failed to %q: %v", endpoint, err)
	}
	conn.Close()
	return nil
}
