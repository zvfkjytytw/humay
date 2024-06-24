package humaystorage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const (
	existTableQuery         = `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = '%s') AS table_exist`
	createMetricTypeQuery   = "CREATE TYPE metric_type AS ENUM ('gauge', 'counter')"
	createGaugeTableQuery   = "CREATE TABLE gauge_metrics(name text NOT NULL UNIQUE, value double precision NOT NULL)"
	createCounterTableQuery = "CREATE TABLE counter_metrics (name text NOT NULL UNIQUE, value bigint NOT NULL)"
)

func (s *PGStorage) initDB() error {
	if !(s.checkTableExist(counterTable) || s.checkTableExist(gaugeTable)) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tx, err := s.dbConnect.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return err
		}

		initCommands := []string{
			// createMetricTypeQuery,
			createGaugeTableQuery,
			createCounterTableQuery,
		}

		s.checkTableExist(counterTable)

		for _, command := range initCommands {
			_, err := tx.Exec(command)
			if err != nil {
				_ = tx.Rollback()
				return err
			}
		}

		if err = tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

func (s *PGStorage) checkTableExist(table string) bool {
	var result bool
	row := s.dbConnect.QueryRow(fmt.Sprintf(existTableQuery, table))
	row.Scan(&result)

	return result
}
