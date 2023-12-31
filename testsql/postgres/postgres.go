// Package postgres contains the basic definition for connection to a postgres
// based container
package postgres

import (
	"context"
	"fmt"
	"strconv"
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
}

func NewDefaultConnector(ctx context.Context) *Connector {
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

	return c
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

func (c *Connector) PortWithProtocol() string {
	return strconv.FormatInt(int64(c.port), 10) + "/tcp"
}

func (c *Connector) Env() []string {
	return []string{"POSTGRES_PASSWORD=" + c.databasePassword, "POSTGRES_USER=" + c.databaseUser, "POSTGRES_DB=" + c.databaseName}
}
