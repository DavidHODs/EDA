package main

import (
	"github.com/DavidHODs/EDA/natsMessage"
)

func main() {
	// app := fiber.New()
	// app.Get("/", natsMessage.NatsOpsDemo)
	// app.Listen(":3000")
	// database.InitDB()
	natsMessage.NatsOps()
}
