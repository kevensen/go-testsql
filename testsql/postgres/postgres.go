package postgres

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"testing"

	"github.com/cenkalti/backoff"
	"github.com/docker/docker/api/types/container"
	tsqlcon "github.com/kevensen/go-testsql/testsql/container"

	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	defaultPort          = 5432
	defaultDB            = "postgres"
	defaultUser          = "postgres"
	defaultPassword      = "postgres"
	defaultImage         = "postgres:13"
	defaultHost          = "localhost"
	defaultContainerName = "go-testsql-postgres"
)

type Connector struct {
	port             int
	databaseName     string
	databaseUser     string
	databasePassword string
	databaseHost     string
	additionalArgs   map[string]string
	containerImage   string
	dockerClient     *client.Client
	containerID      string
}

func NewDefaultConnector(ctx context.Context, t *testing.T) (*Connector, func()) {
	t.Helper()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	c := &Connector{
		port:             defaultPort,
		databaseName:     defaultDB,
		databaseUser:     defaultUser,
		databasePassword: defaultPassword,
		containerImage:   defaultImage,
		dockerClient:     cli,
		additionalArgs: map[string]string{
			"sslmode": "disable",
		},
	}

	c.containerID = tsqlcon.StartContainer(ctx, t, cli, c.ContainerConfig(), c.containerImage, defaultContainerName)
	c.databaseHost = tsqlcon.ContainerIP(ctx, t, c.dockerClient, c.containerID)

	operation := func() error {
		return c.testConnection() // or an error
	}

	err = backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		err := c.dockerClient.ContainerStop(ctx, c.containerID, container.StopOptions{})
		if err != nil {
			t.Fatal(err)
		}
	}

	return c, cleanup
}

func (c *Connector) ContainerConfig() *container.Config {
	return &container.Config{
		Image:        c.containerImage,
		Tty:          false,
		ExposedPorts: nat.PortSet{nat.Port(strconv.FormatInt(int64(c.port), 10)): struct{}{}},
		Env:          []string{"POSTGRES_PASSWORD=" + c.databasePassword, "POSTGRES_USER=" + c.databaseUser, "POSTGRES_DB=" + c.databaseName},
	}
}

// DataSourceName returns the data source name expected by https://github.com/jackc/pgx
func (c *Connector) DataSourceName() string {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", c.databaseHost, c.databaseUser, c.databasePassword, c.databaseName, c.port)
	for k, v := range c.additionalArgs {
		dsn += fmt.Sprintf(" %s=%s", k, v)
	}
	return dsn
}

func (c *Connector) testConnection() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.databaseHost, c.port))
	if err != nil {
		return err

	}
	conn.Close()
	return nil
}
