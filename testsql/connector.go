// Package testsql contains the primary mechanism for connecting to a database
// container for unit or other functional testing.
package testsql

import (
	"context"
	"fmt"
	"net"
	"testing"

	tsqlcon "github.com/kevensen/go-testsql/testsql/container"

	"github.com/cenkalti/backoff"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Database interface {
	ContainerConfig() *container.Config
	HostConfig() *container.HostConfig
	DataSourceName(string) string
	ContainerImage() string
	ContainerName() string
	Port() int
}

type TestConnector struct {
	host   string
	dbConn Database
}

func New(ctx context.Context, t *testing.T, dbConn Database) (*TestConnector, func()) {
	t.Helper()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatal(err)
	}

	testConn := &TestConnector{
		dbConn: dbConn,
	}

	containerID, err := tsqlcon.StartContainer(ctx, cli, dbConn.ContainerConfig(), dbConn.HostConfig(), dbConn.ContainerImage(), dbConn.ContainerName())
	if err != nil {
		t.Fatal(err)
	}
	testConn.host, err = tsqlcon.ContainerIP(ctx, cli, containerID)
	if err != nil {
		t.Fatal(err)
	}

	operation := func() error {
		return testConnection(testConn.host, dbConn.Port()) // or an error
	}

	err = backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		err := cli.ContainerStop(ctx, containerID, container.StopOptions{})
		if err != nil {
			t.Fatal(err)
		}
		cli.Close()
	}
	return testConn, cleanup

}

func (tc *TestConnector) DataSourceName() string {
	return tc.dbConn.DataSourceName(tc.host)
}

func testConnection(host string, port int) error {
	endpoint := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", endpoint)
	if err != nil {
		return fmt.Errorf("connection test failed to %q: %v", endpoint, err)

	}
	conn.Close()
	return nil
}
