package postgresdb

import (
	"context"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)


type postgresContainer struct {
	timeout       time.Duration
	sslmode       string
	dbName        string
	user          string
	password      string
	migrationPath string
}

func DefaultPostgresContainer() *postgresContainer {
	return &postgresContainer{
		timeout:  10 * time.Second,
		dbName:   "postgres",
		user:     "postgres",
		password: "postgres",
		sslmode:  "disable",
	}
}

func (c *postgresContainer) WithTimeout(d time.Duration) {
	c.timeout = d
}

func (c *postgresContainer) WithDbUser(u string) {
	c.user = u
}

func (c *postgresContainer) WithDbPassword(p string) {
	c.password = p
}

func (c *postgresContainer) WithDbName(db string) {
	c.dbName = db
}

func (c *postgresContainer) WithDbSSLMode(mode string) {
	c.sslmode = mode
}


func startPostgres(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:10-alpine",
		ExposedPorts: []string{"5432"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		Env: map[string]string{
			"POSTGRES_PASSWORD": "postgres",
		},
	}

	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}
