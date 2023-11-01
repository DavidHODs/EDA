package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/DavidHODs/EDA/utils"

	_ "github.com/lib/pq"
)

// import "github.com/DavidHODs/EDA/utils"

const (
	logFile = "dbLog.log"
)

func InitDB() *sql.DB {
	logger, err := utils.Logger(logFile)
	if err != nil {
		log.Fatalf("error: could not create database log file: %s", err)
		return nil
	}
	defer logger.Close()

	log.SetOutput(logger)

	envValues, err := utils.LoadEnv("DB_USERNAME", "DB_PASSWORD", "HOST", "DATABASE")
	if err != nil {
		log.Fatalf("error: could not load database environment values: %s", err)
	}
	dbUsername, dbPassword, host, database := envValues[0], envValues[1], envValues[2], envValues[3]

	connStr := fmt.Sprintf("postgresql://%s:%s@%s/%s", dbUsername, dbPassword, host, database)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("error: could not connect to postgres: %s", err)
		return nil
	}

	return db
}
