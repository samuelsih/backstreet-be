package repo

import (
	"backstreetlinkv2/db"
	"backstreetlinkv2/db/migrations"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *sql.DB
var testCache *Cache

func TestMain(m *testing.M) {
	var err error

	cleanup, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}

	testCache, err = NewCache(context.Background(), 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	if err := cleanup(); err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

func setupDB() (func() error, error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mysql:latest",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env: map[string]string{
			"MYSQL_DB":       "backstreet",
			"MYSQL_PASSWORD": "password",
			"MYSQL_USER":     "mysql",
		},
	}

	container, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)

	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("mysql:password@%v:%v/backstreet?tls=skip-verify", hostIP, mappedPort.Port())

	testDB, err = db.ConnectMySQL(uri)
	if testDB == nil {
		return nil, err
	}

	_, err = testDB.ExecContext(ctx, migrations.UpCmd)
	if err != nil {
		log.Fatalf("cant create table: %v", err)
	}

	cleanup := func() error {
		return container.Terminate(ctx)
	}

	return cleanup, nil
}
