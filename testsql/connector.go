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
	defer cli.Close()

	testConn := &TestConnector{
		dbConn: dbConn,
	}

	containerID := tsqlcon.StartContainer(ctx, t, cli, dbConn.ContainerConfig(), dbConn.ContainerImage(), dbConn.ContainerName())
	testConn.host = tsqlcon.ContainerIP(ctx, t, cli, containerID)

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
	}
	return testConn, cleanup

}

func (tc *TestConnector) DataSourceName() string {
	return tc.dbConn.DataSourceName(tc.host)
}

func testConnection(host string, port int) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err

	}
	conn.Close()
	return nil
}
