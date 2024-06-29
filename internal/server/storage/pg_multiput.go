package humaystorage

import (
	"database/sql"
	"fmt"
	"strings"
)

const (
	insertQuery = `
	INSERT INTO %s
    (name, value)
	VALUES %s;
	`
	updateQuery = `
	UPDATE %s AS old
	SET value = new.value
	FROM (VALUES %s)
	AS new(name, value)
	WHERE old.name = new.name;
	`
)

var valueType = map[string]string{
	counterTable: "bigint",
	gaugeTable:   "float8",
}

func (s *PGStorage) PutGaugeMetrics(metrics map[string]float64) (err error) {
	forUpdate := make(map[string]float64)
	forInsert := make(map[string]float64)

	for name, value := range metrics {
		if _, err = s.GetGaugeMetric(name); err != nil {
			forInsert[name] = value
		} else {
			forUpdate[name] = value
		}
	}

	// Update metrics
	if len(forUpdate) > 0 {
		if err = putMetrics(s.dbConnect, gaugeTable, updateQuery, forUpdate); err != nil {
			return fmt.Errorf("failed update gauge metrics: %v", err)
		}
	}

	// Insert metrics
	if len(forInsert) > 0 {
		if err = putMetrics(s.dbConnect, gaugeTable, insertQuery, forInsert); err != nil {
			return fmt.Errorf("failed insert gauge metrics: %v", err)
		}
	}

	return nil
}

func (s *PGStorage) PutCounterMetrics(metrics map[string]int64) (err error) {
	forUpdate := make(map[string]int64)
	forInsert := make(map[string]int64)

	for name, delta := range metrics {
		if value, err := s.GetCounterMetric(name); err != nil {
			forInsert[name] = delta
		} else {
			forUpdate[name] = value + delta
		}
	}

	// Update metrics
	if len(forUpdate) > 0 {
		if err = putMetrics(s.dbConnect, counterTable, updateQuery, forUpdate); err != nil {
			return fmt.Errorf("failed update counter metrics: %v", err)
		}
	}

	// Insert metrics
	if len(forInsert) > 0 {
		if err = putMetrics(s.dbConnect, counterTable, insertQuery, forInsert); err != nil {
			return fmt.Errorf("failed insert counter metrics: %v", err)
		}
	}

	return nil
}

func putMetrics[T Number](db *sql.DB, table, query string, metrics map[string]T) error {
	i := 1
	metricsLen := len(metrics)
	args := make([]any, 0, 2*metricsLen)
	values := make([]string, 0, metricsLen)

	for name, value := range metrics {
		values = append(values, fmt.Sprintf("($%d, $%d::%s)", i, i+1, valueType[table]))
		args = append(args, name, value)
		i += 2
	}
	sql := fmt.Sprintf(query, table, strings.Join(values, ","))

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed init DB transaction: %v", err)
	}

	result, err := tx.Exec(sql, args...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed execute query: %v", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed get count of the affected rows: %v", err)
	}
	if n != int64(metricsLen) {
		tx.Rollback()
		return fmt.Errorf("affected %d rows instead %d", n, metricsLen)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed commit query result: %v", err)
	}

	return nil
}
