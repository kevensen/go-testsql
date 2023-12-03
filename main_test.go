package main

import (
	"context"
	"os"
	"testing"

	tsqlp "github.com/kevensen/go-testsql/testsql"
	psqltest "github.com/kevensen/go-testsql/testsql/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func TestSomeSQL(t *testing.T) {
	ctx := context.Background()
	testDB, cleanup := tsqlp.New(ctx, t, psqltest.NewDefaultConnector(ctx))
	defer cleanup()
	var err error
	db, err = gorm.Open(postgres.Open(testDB.DataSourceName()), &gorm.Config{})

	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
