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
	var initCommands []string

	if !s.checkTableExist(counterTable) {
		initCommands = append(initCommands, createCounterTableQuery)
	}

	if !s.checkTableExist(gaugeTable) {
		initCommands = append(initCommands, createGaugeTableQuery)
	}

	if len(initCommands) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tx, err := s.dbConnect.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return fmt.Errorf("failed begin init transaction: %v", err)
		}

		for _, command := range initCommands {
			_, err := tx.Exec(command)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed init database: %v", err)
			}
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed commit init transaction: %v", err)
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
