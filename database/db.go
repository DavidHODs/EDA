package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/DavidHODs/EDA/utils"
	"github.com/rs/zerolog"

	_ "github.com/lib/pq"
)

// InitDB establishes a connection with postgres database
func InitDB(log *os.File) *sql.DB {
	// logger uses Go time layout for time stamping
	zerolog.TimeFieldFormat = zerolog.TimestampFunc().Format("2006-01-02T15:04:05Z07:00")

	// sets up a logger with the specified log file as the log output destination
	logger := zerolog.New(log).With().Timestamp().Caller().Logger()

	envValues, err := utils.LoadEnv("DB_USERNAME", "DB_PASSWORD", "HOST", "DATABASE")
	if err != nil {
		logger.Fatal().
			Str("error", "utility error").
			Msg("could not load env file")

	}
	dbUsername, dbPassword, host, database := envValues[0], envValues[1], envValues[2], envValues[3]

	connStr := fmt.Sprintf("postgresql://%s:%s@%s/%s", dbUsername, dbPassword, host, database)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatal().
			Str("error", "pq error").
			Msgf("could not connect to postgres: %s", err)
		return nil
	}

	// Pings the database to verify if connection is alive.
	err = db.Ping()
	if err != nil {
		logger.Fatal().
			Str("error", "database error").
			Msgf("ping error: %s", err)
		return nil
	}

	return db
}
