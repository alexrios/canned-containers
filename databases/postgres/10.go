package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const connStrTemplate string = "user=%s dbname=%s host=%s port=%s password=%s sslmode=%s"
const defaultDBScenarioPath = "file://."

type containerizedDatabaseContext struct {
	Conn      *sql.DB
	Container testcontainers.Container
	Ctx       context.Context
}

type containerDatabaseTest struct {
	timeout       time.Duration
	sslmode       string
	dbName        string
	user          string
	password      string
	migrationPath string
}

func DefaultContainerDatabaseTest() *containerDatabaseTest {
	return &containerDatabaseTest{
		timeout:       10 * time.Second,
		dbName:        "postgres",
		user:          "postgres",
		password:      "postgres",
		sslmode:       "disable",
		migrationPath: defaultDBScenarioPath,
	}
}

func (c *containerDatabaseTest) WithTimeout(d time.Duration) {
	c.timeout = d
}

func (c *containerDatabaseTest) WithDbUser(u string) {
	c.user = u
}

func (c *containerDatabaseTest) WithDbPassword(p string) {
	c.password = p
}

func (c *containerDatabaseTest) WithDbName(db string) {
	c.dbName = db
}

func (c *containerDatabaseTest) WithDbSSLMode(mode string) {
	c.sslmode = mode
}

func (c *containerDatabaseTest) WithMigrationFile(path string) {
	c.migrationPath = fmt.Sprintf("file://%s", path)
}

func (cdt *containerDatabaseTest) Setup(t *testing.T) (*sql.DB, func()) {
	dbCtx := createContainerContext(cdt, t)
	return dbCtx.Conn, func() {
		if dbCtx.Conn != nil {
			_ = dbCtx.Conn.Close()
		}
		_ = dbCtx.Container.Terminate(dbCtx.Ctx)
	}
}

func createContainerContext(cdt *containerDatabaseTest, t *testing.T) *containerizedDatabaseContext {
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(cdt.timeout, cancel)
	postgresC, err := startPostgres(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	connStr, err := buildConnStringFromContainer(ctx, cdt, postgresC)
	if err != nil {
		t.Fatal(err.Error())
	}
	//GETTING DB CONNECTION
	db := connect(ctx, connStr, t)
	// if !strings.HasSuffix(err.Error(), "connection reset by peer") ||
	// 	!strings.HasSuffix(err.Error(), "the database system is starting up") {
	// 	t.Fatalf("could not ping DB... %v", err)
	// }
	//RUN MIGRATIONS
	err = runMigrations(cdt.migrationPath, db, t)
	if err != nil {
		t.Fatal(err.Error())
	}

	databaseContext := &containerizedDatabaseContext{
		Conn:      db,
		Container: postgresC,
	}
	return databaseContext
}

func connect(ctx context.Context, connStr string, t *testing.T) *sql.DB {
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("could not connect to the Postgres database... %v", err)
	}
	ping := make(chan bool)

	go func(dbConn *sql.DB, c chan bool) {
		defer close(ping)
		for {
			err = dbConn.Ping()
			if err == nil {
				c <- true
				break
			}
		}
	}(dbConn, ping)

	select {
	case <-ping:
		return dbConn
	case <-ctx.Done():
		t.Fatalf("could not ping DB... %v", err)
	}

	return dbConn
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

func runMigrations(migrationsPath string, dbConn *sql.DB, t *testing.T) error {
	driver, err := postgres.WithInstance(dbConn, &postgres.Config{})
	if err != nil {
		t.Fatalf("migration failed... %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath, "postgres", driver)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if m != nil {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			t.Fatalf("An error occurred while syncing the database.. %v", err)
		}
	}
	return err
}

func buildConnStringFromContainer(ctx context.Context, cdt *containerDatabaseTest, postgresC testcontainers.Container) (string, error) {
	host, err := postgresC.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(connStrTemplate, cdt.user, cdt.dbName, host, port.Port(), cdt.password, cdt.sslmode), nil
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
