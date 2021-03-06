package postgres_example

import (
	"context"
	"github.com/alexrios/canned-containers/databases/postgres"
	_ "github.com/lib/pq"
	"testing"
	"time"
)

func TestPostgres(t *testing.T) {
	t.Run("start a new Postgres container", func(t *testing.T) {
		container := postgresdb.DefaultPostgresContainer()
		container.WithTimeout(1 * time.Minute)
		databaseContext, err := container.CreateContainerContext()
		if err != nil {
			t.Fatal(err)
		}
		if databaseContext.Conn == nil {
			t.FailNow()
		}
		//Paranoia check
		err = databaseContext.Conn.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = databaseContext.Container.Terminate(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	})
}
