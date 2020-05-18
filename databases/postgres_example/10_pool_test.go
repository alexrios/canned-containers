package postgrtes_test

import (
	"context"
	"github.com/alexrios/canned-containers/databases/postgres"
	_ "github.com/lib/pq"
	"testing"
	"time"
)

func TestPoolPostgres(t *testing.T) {
	t.Run("start a new Postgres Pool container", func(t *testing.T) {
		container := postgresdb.DefaultPostgresContainer()
		container.WithTimeout(1 * time.Minute)
		databaseContext, err := container.CreatePoolContainerContext()
		if err != nil {
			t.Fatal(err)
		}
		if databaseContext.Pool == nil {
			t.FailNow()
		}
		//Paranoia check
		databaseContext.Pool.Close()

		err = databaseContext.Container.Terminate(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	})
}
