// Package mockseed provides functionality to seed mock data into the database for testing purposes.
package mockseed

import (
	"context"
	"database/sql"

	gl "github.com/kubex-ecosystem/logz"
	_ "github.com/lib/pq" // se for postgres, por exemplo
)

func SeedMockData(ctx context.Context, db *sql.DB) error {
	gl.Info("Seed mock data is not implemented for Golang yet.")
	return nil
}
