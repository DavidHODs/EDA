package natsMessage

import (
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/DavidHODs/EDA/database"
	"github.com/DavidHODs/EDA/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/nats-io/nats.go"
	"golang.org/x/net/context"
)

const (
	listenerOneSubject   = "events.chain"
	listenerTwoSubject   = "events.chain.listener2"
	listenerThreeSubject = "events.chain.listener3"
	logFile              = "natsLog.log"
)

type EventRequest struct {
	EventName string `json:"eventName"`
}

type EventResponse struct {
	Listener1 string    `json:"listener1" db:"listener_one"`
	Listener2 string    `json:"listener2" db:"listener_two"`
	Listener3 string    `json:"listener3" db:"listener_three"`
	EventTime time.Time `json:"eventTime" db:"event_time"`
}

// forwardMessage publishes a payload modified by the respective listener
func forwardMessage(nc *nats.Conn, subject, forwardingListener string, data []byte) {
	err := nc.Publish(subject, data)
	if err != nil {
		log.Printf("%s: error publishing message: %v", forwardingListener, err)
	}
}

// NatsOpsDemo connects with NATS server, subscribes, and publishes modified payload - over HTTP Request
func NatsOpsDemo(c *fiber.Ctx) error {
	var eventReq EventRequest
	err := c.BodyParser(&eventReq)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Printf("could not connect to NATS server: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	defer nc.Close()

	var listenerOneData string

	// Create a subscriber to receive the message.
	sub, err := nc.Subscribe(listenerOneSubject, func(msg *nats.Msg) {
		listenerOneData = string(msg.Data)
		log.Println(listenerOneData)
	})
	if err != nil {
		log.Printf("error: listener1 could not subscribe: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	defer sub.Unsubscribe()

	// Publish the message after subscribing.
	err = nc.Publish(listenerOneSubject, []byte(eventReq.EventName))
	if err != nil {
		log.Printf("error publishing message: %s", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": listenerOneData})
}

// NatsOps connects with nats server, subscribes and publishes modified payload
func NatsOps() {
	logger, err := utils.Logger(logFile)
	if err != nil {
		log.Fatalf("error: could not create nats ops log file: %s", err)
	}
	defer logger.Close()

	log.SetOutput(logger)
	// retrieves nats url from env file
	envValue, err := utils.LoadEnv("NATS_URL")
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	url := envValue[0]

	// connects to nats server using the specified url
	nc, err := nats.Connect(url)
	if err != nil {
		log.Fatalf("could not connect to nats server: %s", err)
	}
	defer nc.Close()

	// creates listener variables to store information
	var listenerOne, listenerTwo, listenerThree *string

	done := make(chan bool)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// creates a wait group to ensure all neccessary data are stored before an attempt is made to lg or store them
	var wg sync.WaitGroup
	wg.Add(3)

	// creates the first listener that receives a payload from the nats cli
	sub1, err := nc.Subscribe(listenerOneSubject, func(msg *nats.Msg) {
		defer wg.Done()

		listenerOneData := string(msg.Data)
		listenerOne = &listenerOneData
		log.Printf("listener1 received %s ", listenerOneData)

		// converts received data to all caps
		capitalizedData := []byte(strings.ToUpper(listenerOneData))

		// publishes modified data for second listener to pick up
		forwardMessage(nc, listenerTwoSubject, "listener1", capitalizedData)
	})
	if err != nil {
		log.Printf("error: listener1 could not subscribe: %v", err)
	}
	defer sub1.Drain()

	// creates the second listener that receives a payload from the first listener
	sub2, err := nc.Subscribe(listenerTwoSubject, func(msg *nats.Msg) {
		defer wg.Done()

		listenerTwoData := string(msg.Data)
		listenerTwo = &listenerTwoData
		log.Printf("listener2 received %s from listener1", listenerTwoData)

		// reverses received data
		reversedData := []byte(utils.ReverseString(listenerTwoData))

		// publishes modified data for third listener to pick up
		forwardMessage(nc, listenerThreeSubject, "listener2", reversedData)
	})
	if err != nil {
		log.Printf("error: listener2 could not subscribe: %v", err)
	}
	defer sub2.Drain()

	// creates the third listener that receives a payload from the second listener
	sub3, err := nc.Subscribe(listenerThreeSubject, func(msg *nats.Msg) {
		defer wg.Done()

		listenerThreeData := string(msg.Data)
		log.Printf("listener3 received %s from listener2", listenerThreeData)

		// converts received data to lower form
		lowerDataStr := strings.ToLower(listenerThreeData)
		lowerData := []byte(strings.ToLower(listenerThreeData))

		log.Printf("listener3 modified %s from listener2 into %s", listenerThreeData, lowerData)
		listenerThree = &lowerDataStr

		// *** the following lines are commented out - it triggers a negative waitgroup error plus it creates an infinite loop of multiple subscribers and publishers actions ***

		// forwardMessage(nc, listenerOneSubject, "listener3", lowerData)
		// forwardMessage(nc, listenerTwoSubject, "listener3", lowerData)
	})
	if err != nil {
		log.Printf("error: listener3 could not subscribe: %v", err)
	}
	defer sub3.Drain()

	// this goroutine checks for context timeout (guards against process running forever) and also watches for the done channel, an indication that all required processes are done runnning before exiting the program. runtime.GoExit allows all deferred function to get called before exiting
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Fatal("operation cancelled: operation took too long")

			case <-done:
				log.Println("processes succsessfully carried out")
				runtime.Goexit()
			}
		}
	}()

	// blocks this part of the code until all subscribers have gotten their data, guards against runtime panic: nil dereference error
	wg.Wait()

	eventResp := EventResponse{
		Listener1: *listenerOne,
		Listener2: *listenerTwo,
		Listener3: *listenerThree,
		EventTime: time.Now(),
	}

	db := database.InitDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO events (listener_one, listener_two, listener_three, event_time) VALUES ($1, $2, $3, $4)")
	if err != nil {
		log.Printf("error: could not prepare database tranasaction statement: %s", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(eventResp.Listener1, eventResp.Listener2, eventResp.Listener3, eventResp.EventTime)
	if err != nil {
		log.Printf("error: could not insert event record into database: %s", err)
	}

	log.Println(eventResp)

	// closed done channel indicates end of required processes
	close(done)
}
