package main

import (
	"context"
	"flag"
	"os"
	"strconv"
	"strings"

	serverApp "github.com/zvfkjytytw/humay/internal/server/app"
	humayHTTPServer "github.com/zvfkjytytw/humay/internal/server/http"
)

const (
	addressEnv         = "ADDRESS"
	restoreEnv         = "RESTORE"
	storageIntervalEnv = "STORE_INTERVAL"
	fileStoragePathEnv = "FILE_STORAGE_PATH"
	databaseDSNEnv     = "DATABASE_DSN"
)

func main() {
	var (
		// Server config file
		configFile string
		// Server address and port
		address string
		// interval for saving data
		storageInterval int
		// file for save/restore date
		fileStoragePath string
		// flag for restore data
		restore bool
		// dsn for connect to postgres
		databaseDSN string
	)

	flag.StringVar(&configFile, "c", "./build/server.yaml", "Server config file")
	flag.StringVar(&address, "a", "localhost:8080", "Server address")
	flag.StringVar(&fileStoragePath, "f", "/tmp/metrics-db.json", "Data storage file")
	flag.IntVar(&storageInterval, "i", 300, "Interval between saving data")
	flag.BoolVar(&restore, "r", true, "Restore data at the time of launch")
	flag.StringVar(&databaseDSN, "d", "", "DSN for postgreSQL connection")
	flag.Parse()

	value, ok := os.LookupEnv(addressEnv)
	if ok {
		address = value
	}
	host, port := splitAddress(address)

	value, ok = os.LookupEnv(databaseDSNEnv)
	if ok {
		databaseDSN = value
	}

	saverConfig, err := getSaverConfig(storageInterval, fileStoragePath, restore)
	if err != nil {
		panic(err)
	}

	config := &serverApp.ServerConfig{
		HTTPConfig: &humayHTTPServer.HTTPConfig{
			Host:         host,
			Port:         port,
			ReadTimeout:  5,
			WriteTimeout: 10,
			IdleTimeout:  20,
		},
		SaverConfig: saverConfig,
		DatabaseDSN: databaseDSN,
	}

	app, err := serverApp.NewApp(config)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app.Run(ctx)
}

func splitAddress(address string) (host string, port int32) {
	values := strings.Split(address, ":")
	if len(values) == 1 {
		host = values[0]
		port = 8080
		return
	}

	host = values[0]
	p, err := strconv.Atoi(values[1])
	if err != nil {
		port = 8080
	}
	port = int32(p)

	return
}

func getSaverConfig(i int, f string, r bool) (*serverApp.SaverConfig, error) {
	var err error

	sInterval := i
	envI, ok := os.LookupEnv(storageIntervalEnv)
	if ok {
		sInterval, err = strconv.Atoi(envI)
		if err != nil {
			return nil, err
		}
	}

	fStorage := f
	envF, ok := os.LookupEnv(fileStoragePathEnv)
	if ok {
		fStorage = envF
	}

	restore := r
	envR, ok := os.LookupEnv(restoreEnv)
	if ok {
		restore, err = strconv.ParseBool(envR)
		if err != nil {
			return nil, err
		}
	}

	return &serverApp.SaverConfig{
		Interval:    int32(sInterval),
		StorageFile: fStorage,
		Restore:     restore,
	}, nil
}
