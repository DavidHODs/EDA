package main

import (
	"log"

	"github.com/DavidHODs/EDA/database"
	"github.com/DavidHODs/EDA/natsMessage"
	"github.com/DavidHODs/EDA/utils"
	"github.com/gofiber/fiber/v2"
)

const (
	natsLog = "natsLog.log"
	dbLog   = "dbLog.log"
)

func main() {
	// initializes nats log file connection
	natsLogF, err := utils.Logger(natsLog)
	if err != nil {
		log.Fatalf("error: could not create nats ops log file: %s", err)
	}
	defer natsLogF.Close()

	// initializes db log file connection
	dbLogF, err := utils.Logger(dbLog)
	if err != nil {
		log.Fatalf("error: could not create db log file: %s", err)
	}
	defer dbLogF.Close()

	// initializes nats server connection
	nc := natsMessage.NatServerConn(natsLogF)

	// initializes database connection
	db := database.InitDB(dbLogF)

	// loads ConnectionManager struct with required services and file connections
	cm := &natsMessage.ConnectionManager{
		NC:      nc,
		DB:      db,
		NatsLog: natsLogF,
	}

	app := fiber.New()

	app.Post("/publish", cm.NatsOps)

	app.Listen(":3000")
}
