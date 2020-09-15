package postgresdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"time"
)

const connStrPoolTemplate string = "user=%s dbname=%s host=%s port=%s password=%s pool_max_conns=1 sslmode=%s"

type containerizedDatabasePoolContext struct {
	Pool      *pgxpool.Pool
	Container testcontainers.Container
	Ctx       context.Context
	ConnStr   string
}

func (pc *postgresContainer) CreatePoolContainerContext() (*containerizedDatabasePoolContext, error) {
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(pc.timeout, cancel)
	postgresC, err := startPostgres(ctx)
	if err != nil {
		return nil, err
	}
	connStr, err := buildConnStringFromPoolContainer(ctx, pc, postgresC)
	if err != nil {
		return nil, err
	}

	pool, err := connectPool(ctx, connStr)
	if err != nil {
		return nil, err
	}

	databaseContext := &containerizedDatabasePoolContext{
		Pool:      pool,
		Container: postgresC,
		ConnStr:   connStr,
	}
	return databaseContext, nil
}

func connectPool(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	connChan := make(chan *pgxpool.Pool)
	go tryConnect(ctx, connStr, connChan)

	select {
	case pool := <-connChan:
		return pool, nil
	case <-ctx.Done():
		return nil, errors.New("could not connect to DB")
	}
}

func tryConnect(ctx context.Context, connStr string, c chan *pgxpool.Pool) {
	for {
		pool, err := pgxpool.Connect(ctx, connStr)
		if err == nil {
			c <- pool
			break
		}
	}
}

func buildConnStringFromPoolContainer(ctx context.Context, pc *postgresContainer, postgresC testcontainers.Container) (string, error) {
	host, err := postgresC.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(connStrPoolTemplate, pc.user, pc.dbName, host, port.Port(), pc.password, pc.sslmode), nil
}
