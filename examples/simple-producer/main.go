package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/streadway/amqp"
)

type Config struct {
	AMQP struct {
		Addr string `default:"amqp://guest:guest@rabbit:5672"`
	}

	App struct {
		Exchange   string        `default:"messages"`
		RoutingKey string        `default:"message.published"`
		Messages   int           `default:"10"`
		Sleep      time.Duration `default:"1s"`
	}
}

var wg sync.WaitGroup

func main() {
	logger := log.New(os.Stderr, "[go-carrot/simple-producer] ", log.LstdFlags|log.Lshortfile)

	var config Config
	mustNotFail(envconfig.Process("", &config), logger)

	conn, err := amqp.Dial(config.AMQP.Addr)
	mustNotFail(err, logger)
	defer conn.Close()

	ch, err := conn.Channel()
	mustNotFail(err, logger)
	defer ch.Close()

	start := time.Now()

	logger.Printf("Producing on '%s' exchange, '%s' routing key...\n",
		config.App.Exchange,
		config.App.RoutingKey,
	)

	for i := 0; i < config.App.Messages; i++ {
		mustNotFail(ch.Publish(
			config.App.Exchange,
			config.App.RoutingKey,
			false,
			false,
			amqp.Publishing{
				MessageId: fmt.Sprintf("%d", rand.Int63()),
				Body:      []byte(fmt.Sprintf("message %d", i)),
			},
		), logger)

		<-time.After(config.App.Sleep)
	}

	logger.Printf("Produced %d messages after %s. See ya!\n",
		config.App.Messages,
		time.Since(start),
	)
}

func mustNotFail(err error, logger *log.Logger) {
	if err != nil {
		logger.Fatalln("Unexpected error:", err)
	}
}
