package main

import (
	"context"
	"os"
	"testing"

	tsqlp "github.com/kevensen/go-testsql/testsql/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func TestSomeSQL(t *testing.T) {
	ctx := context.Background()
	postgresConnector, cleanup := tsqlp.NewDefaultConnector(ctx, t)
	defer cleanup()
	var err error
	db, err = gorm.Open(postgres.Open(postgresConnector.DataSourceName()), &gorm.Config{})

	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
