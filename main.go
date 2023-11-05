// package main

// import (
// 	"github.com/DavidHODs/EDA/natsMessage"
// 	"github.com/gofiber/fiber/v2"
// )

// func main() {
// 	app := fiber.New()
// 	app.Post("/", natsMessage.NatsOpsDemo)
// 	app.Listen(":3000")
// 	// natsMessage.NatsOps()
// }

package main

import (
	"github.com/DavidHODs/EDA/natsMessage"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Post("/publish", natsMessage.NatsOps)

	app.Listen(":3000")
}
