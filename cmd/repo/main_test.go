package repo

import (
	"backstreetlinkv2/db"
	"backstreetlinkv2/db/migrations"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"os"
	"testing"
)

var testDB *pgx.Conn

func TestMain(m *testing.M) {
	var err error

	cleanup, err := setup()
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	if err := cleanup(); err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

func setup() (func() error, error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env: map[string]string{
			"POSTGRES_DB":       "backstreet",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_USER":     "postgres",
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

	uri := fmt.Sprintf("postgres://postgres:postgres@%v:%v/backstreet?sslmode=disable", hostIP, mappedPort.Port())

	testDB, err = db.ConnectPG(uri)
	if testDB == nil {
		return nil, err
	}

	_, err = testDB.Exec(ctx, migrations.UpCmd)
	if err != nil {
		log.Fatalf("cant create table: %v", err)
	}

	cleanup := func() error {
		return container.Terminate(ctx)
	}

	return cleanup, nil
}
