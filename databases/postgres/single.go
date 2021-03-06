package postgresdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

const connStrTemplate string = "postgres://%s:%s@%s:%s/%s?sslmode=%s"

type containerizedDatabaseContext struct {
	Conn      *sql.DB
	Container testcontainers.Container
	Ctx       context.Context
	ConnStr   string
}

func (pc *postgresContainer) CreateContainerContext() (*containerizedDatabaseContext, error) {
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(pc.timeout, cancel)
	postgresC, err := startPostgres(ctx)
	if err != nil {
		return nil, err
	}
	connStr, err := buildConnStringFromContainer(ctx, pc, postgresC)
	if err != nil {
		return nil, err
	}

	db, err := connect(ctx, connStr)
	if err != nil {
		return nil, err
	}

	databaseContext := &containerizedDatabaseContext{
		Conn:      db,
		Container: postgresC,
		ConnStr:   connStr,
	}
	return databaseContext, nil
}

func connect(ctx context.Context, connStr string) (*sql.DB, error) {
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("could not connect to the Postgres database... %v", err)
	}
	pingChan := make(chan bool)
	go ping(dbConn, pingChan)

	select {
	case <-pingChan:
		return dbConn, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("could not ping DB... %v", err)
	}
}

func ping(dbConn *sql.DB, c chan bool) {
	var err error
	for {
		err = dbConn.Ping()
		if err == nil {
			c <- true
			break
		}
	}
}

func buildConnStringFromContainer(ctx context.Context, pc *postgresContainer, postgresC testcontainers.Container) (string, error) {
	host, err := postgresC.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(connStrTemplate, pc.user, pc.password, host, port.Port(), pc.dbName, pc.sslmode), nil
}
