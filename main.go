package main

import (
	"log"

	"github.com/DavidHODs/EDA/database"
	"github.com/DavidHODs/EDA/natsMessage"
	"github.com/gofiber/fiber/v2"
	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}

	db := database.InitDB()

	cm := &natsMessage.ConnectionManager{
		NC: nc,
		DB: db,
	}

	app := fiber.New()

	app.Post("/publish", cm.NatsOps)

	app.Listen(":3000")
}
