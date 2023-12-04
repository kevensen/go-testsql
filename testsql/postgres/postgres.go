// Package postgres contains the basic definition for connection to a postgres
// based container
package postgres

import (
	"context"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types/container"
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
	additionalArgs   map[string]string
	containerImage   string
	containerName    string
	localhost        bool
}

type PostgresOptions interface{}

type BindToLocalHost bool

func NewDefaultConnector(ctx context.Context, opts ...PostgresOptions) *Connector {
	c := &Connector{
		port:             defaultPort,
		databaseName:     defaultDB,
		databaseUser:     defaultUser,
		databasePassword: defaultPassword,
		containerImage:   defaultImage,
		containerName:    defaultContainerName,
		additionalArgs: map[string]string{
			"sslmode": "disable",
		},
	}

	for _, opt := range opts {
		switch v := opt.(type) {
		case BindToLocalHost:
			c.localhost = bool(v)
		}
	}

	return c
}

func (c *Connector) ContainerConfig() *container.Config {
	return &container.Config{
		Image:        c.containerImage,
		Tty:          false,
		ExposedPorts: nat.PortSet{nat.Port(strconv.FormatInt(int64(c.port), 10)): struct{}{}},
		Env:          []string{"POSTGRES_PASSWORD=" + c.databasePassword, "POSTGRES_USER=" + c.databaseUser, "POSTGRES_DB=" + c.databaseName},
	}
}

func (c *Connector) HostConfig() *container.HostConfig {

	if c.localhost {
		return &container.HostConfig{
			PortBindings: nat.PortMap{
				"5432/tcp": []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "5432",
					},
				},
			},
		}
	}
	return nil
}

// DataSourceName returns the data source name expected by https://github.com/jackc/pgx
func (c *Connector) DataSourceName(hostname string) string {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", hostname, c.databaseUser, c.databasePassword, c.databaseName, c.port)
	for k, v := range c.additionalArgs {
		dsn += fmt.Sprintf(" %s=%s", k, v)
	}
	return dsn
}

func (c *Connector) ContainerName() string {
	return c.containerName
}

func (c *Connector) ContainerImage() string {
	return c.containerImage
}

func (c *Connector) Port() int {
	return c.port
}
