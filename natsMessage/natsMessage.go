package natsMessage

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/DavidHODs/EDA/database"
	"github.com/DavidHODs/EDA/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
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
func forwardMessage(nc *nats.Conn, subject, forwardingListener string, data []byte) error {
	err := nc.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("%s: error publishing message: %v", forwardingListener, err)
	}

	return nil
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	defer nc.Close()

	var listenerOneData string

	// Create a subscriber to receive the message.
	sub, err := nc.Subscribe(listenerOneSubject, func(msg *nats.Msg) {
		listenerOneData = string(msg.Data)
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	defer sub.Unsubscribe()

	// Publish the message after subscribing.
	err = nc.Publish(listenerOneSubject, []byte(eventReq.EventName))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": listenerOneData})
}

// NatsOps connects with nats server, subscribes and publishes modified payload
func NatsOps() {
	logF, err := utils.Logger(logFile)
	if err != nil {
		log.Fatalf("error: could not create nats ops log file: %s", err)
	}
	defer logF.Close()

	// logger uses Go time layout for time stamping
	zerolog.TimeFieldFormat = zerolog.TimestampFunc().Format("2006-01-02T15:04:05Z07:00")

	// sets up a logger with the created log file as the log output destination
	logger := zerolog.New(logF).With().Timestamp().Caller().Logger()

	// retrieves nats url from env file
	envValue, err := utils.LoadEnv("NATS_URL")
	if err != nil {
		logger.Fatal().
			Str("error", "utility error").
			Msg("could not load env file")
	}
	url := envValue[0]

	// connects to nats server using the specified url
	nc, err := nats.Connect(url)
	if err != nil {
		logger.Fatal().
			Str("nats", "connection error").
			Msg("could not connect to nats server using the specified url")
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
		logger.Info().
			Str("listener", "listener one").
			Msgf("listener 1 received %s", listenerOneData)

		// converts received data to all caps
		capitalizedData := []byte(strings.ToUpper(listenerOneData))

		// publishes modified data for second listener to pick up
		err = forwardMessage(nc, listenerTwoSubject, "listener1", capitalizedData)
		if err != nil {
			logger.Error().
				Str("publisher", "publisher two").
				Msgf("%s", err)
		}
	})
	if err != nil {
		logger.Error().
			Str("listener1", "subscribtion error").
			Msgf("error: listener1 could not subscribe: %v", err)
	}
	defer sub1.Drain()

	// creates the second listener that receives a payload from the first listener
	sub2, err := nc.Subscribe(listenerTwoSubject, func(msg *nats.Msg) {
		defer wg.Done()

		listenerTwoData := string(msg.Data)
		listenerTwo = &listenerTwoData
		logger.Info().
			Str("listener", "listener two").
			Msgf("listener 2 received %s from listener 1", listenerTwoData)

		// reverses received data
		reversedData := []byte(utils.ReverseString(listenerTwoData))

		// publishes modified data for third listener to pick up
		err = forwardMessage(nc, listenerThreeSubject, "listener2", reversedData)
		if err != nil {
			logger.Error().
				Str("publisher", "publisher three").
				Msgf("%s", err)
		}
	})
	if err != nil {
		logger.Error().
			Str("listener2", "subscribtion error").
			Msgf("error: listener2 could not subscribe: %v", err)
	}
	defer sub2.Drain()

	// creates the third listener that receives a payload from the second listener
	sub3, err := nc.Subscribe(listenerThreeSubject, func(msg *nats.Msg) {
		defer wg.Done()

		listenerThreeData := string(msg.Data)
		logger.Info().
			Str("listener", "listener three").
			Msgf("listener 3 received %s from listener 2", listenerThreeData)

		// converts received data to lower form
		lowerDataStr := strings.ToLower(listenerThreeData)
		lowerData := []byte(strings.ToLower(listenerThreeData))

		logger.Info().
			Str("listener", "listener three").
			Msgf("listener 3 modified %s received from listener 2 into %s", listenerThreeData, lowerData)
		listenerThree = &lowerDataStr

		// *** the following lines are commented out - it triggers a negative waitgroup error plus it creates an infinite loop of multiple subscribers and publishers actions ***

		// forwardMessage(nc, listenerOneSubject, "listener3", lowerData)
		// forwardMessage(nc, listenerTwoSubject, "listener3", lowerData)
	})
	if err != nil {
		logger.Error().
			Str("listener3", "subscribtion error").
			Msgf("error: listener3 could not subscribe: %v", err)
	}
	defer sub3.Drain()

	// this goroutine checks for context timeout (guards against process running forever) and also watches for the done channel, an indication that all required processes are done runnning before exiting the program. runtime.GoExit allows all deferred function to get called before exiting
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Fatal().
					Str("ctx", "ctx done").
					Msg("operation cancelled: operation took too long")

			case <-done:
				logger.Info().
					Msg("all processes done, now gracefully exiting")
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
		logger.Error().
			Str("db", "stmt preparation").
			Msgf("error: could not prepare database tranasaction statement: %s", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(eventResp.Listener1, eventResp.Listener2, eventResp.Listener3, eventResp.EventTime)
	if err != nil {
		logger.Error().
			Str("db", "stmt execution").
			Msgf("error: could not event record into database: %s", err)
	}

	logger.Info().
		Str("response", "struct of data received + modifications").
		Msgf("%v", eventResp)

	// closed done channel indicates end of required processes
	close(done)
}
