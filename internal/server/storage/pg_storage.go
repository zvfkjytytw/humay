package humaystorage

import (
	"database/sql"
	"fmt"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"golang.org/x/exp/constraints"
)

const (
	postgresDriver = "postgres"
	counterTable   = "counter_metrics"
	gaugeTable     = "gauge_metrics"
)

type Number interface {
	constraints.Integer | constraints.Float
}

// type PGStorageConfig struct {
// 	Host     string `yaml:"host" json:"host"`
// 	Port     int32  `yaml:"port" json:"port"`
// 	Database string `yaml:"database" json:"database"`
// 	User     string `yaml:"user" json:"user"`
// 	Password string `yaml:"password" json:"password"`
// }

type PGStorage struct {
	storageType string
	dbConnect   *sql.DB
}

func NewPGStorage(dsn string) (*PGStorage, error) {
	db, err := sql.Open(postgresDriver, dsn)
	if err != nil {
		return nil, err
	}

	pgStorage := &PGStorage{
		storageType: postgresDriver,
		dbConnect:   db,
	}

	err = pgStorage.CheckDBConnect()
	if err != nil {
		return nil, err
	}

	err = pgStorage.initDB()
	if err != nil {
		return nil, err
	}

	return pgStorage, nil
}

func (s *PGStorage) CheckDBConnect() error {
	return s.dbConnect.Ping()
}

func (s *PGStorage) GetType() string {
	return s.storageType
}

func (s *PGStorage) Close() error {
	return s.dbConnect.Close()
}

func (s *PGStorage) GetGaugeMetric(name string) (float64, error) {
	sql, args, err := sq.Select("value").From(gaugeTable).Where(sq.Eq{"name": name}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, err
	}
	row := s.dbConnect.QueryRow(sql, args...)
	var value float64
	err = row.Scan(&value)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (s *PGStorage) PutGaugeMetric(name string, value float64) error {
	_, err := s.GetGaugeMetric(name)
	if err != nil {
		return insertMetric(s.dbConnect, gaugeTable, name, value)
	}

	return updateMetric(s.dbConnect, gaugeTable, name, value)
}

func (s *PGStorage) GetCounterMetric(name string) (int64, error) {
	sql, args, err := sq.Select("value").From(counterTable).Where(sq.Eq{"name": name}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, err
	}
	row := s.dbConnect.QueryRow(sql, args...)
	var value int64
	err = row.Scan(&value)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (s *PGStorage) PutCounterMetric(name string, delta int64) error {
	value, err := s.GetCounterMetric(name)
	if err != nil {
		return insertMetric(s.dbConnect, counterTable, name, delta)
	}

	return updateMetric(s.dbConnect, counterTable, name, value+delta)
}

func (s *PGStorage) GetAllMetrics() map[string]map[string]string {
	metrics := make(map[string]map[string]string)

	// Collect gauge metrics.
	metrics["gauges"] = make(map[string]string)
	sql, args, err := sq.Select("name", "value").From(gaugeTable).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil
	}

	rows, err := s.dbConnect.Query(sql, args...)
	if err != nil {
		return nil
	}
	if err = rows.Err(); err != nil {
		return nil
	}

	for rows.Next() {
		var name string
		var value float64
		rows.Scan(&name, &value)
		metrics["gauges"][name] = strconv.FormatFloat(value, 'f', -1, 64)
	}

	// Collect counter metrics.
	metrics["counters"] = make(map[string]string)
	sql, args, err = sq.Select("name", "value").From(counterTable).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil
	}

	rows, err = s.dbConnect.Query(sql, args...)
	if err != nil {
		return nil
	}
	if err = rows.Err(); err != nil {
		return nil
	}

	for rows.Next() {
		var name string
		var value int64
		rows.Scan(&name, &value)
		metrics["counters"][name] = strconv.FormatInt(value, 10)
	}

	return metrics
}

func insertMetric[T Number](db *sql.DB, table, name string, value T) error {
	sql, args, err := sq.Insert(table).Columns("name", "value").Values(name, value).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return err
	}

	result, err := db.Exec(sql, args...)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("affected %d rows", n)
	}

	return nil
}

func updateMetric[T Number](db *sql.DB, table, name string, value T) error {
	sql, args, err := sq.Update(table).Set("value", value).Where(sq.Eq{"name": name}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return err
	}

	result, err := db.Exec(sql, args...)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("affected %d rows", n)
	}

	return nil
}
