// Sources for https://watermill.io/docs/getting-started/
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

var (
	// For this example, we're using just a simple logger implementation,
	// You probably want to ship your own implementation of `watermill.LoggerAdapter`.
	logger = watermill.NewStdLogger(false, false)
)

func main() {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	handler := router.AddHandler(
		"struct_handler",          // handler name, must be unique
		"incoming_messages_topic", // topic from which we will read events
		pubSub,
		"outgoing_messages_topic", // topic to which we will publish events
		pubSub,
		structHandler{}.Handler,
	)

	handler.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
		return func(message *message.Message) ([]*message.Message, error) {
			log.Println("executing handler specific middleware for ", message.UUID)

			return h(message)
		}
	})

	router.AddNoPublisherHandler(
		"print_outgoing_messages",
		"outgoing_messages_topic",
		pubSub,
		printMessages,
	)

	ctx := context.Background()

	go func() {
		if err := router.Run(ctx); err != nil {
			panic(err)
		}
	}()

	// Guarantee router is up
	time.Sleep(3 * time.Second)
	fmt.Printf("router.IsRunning: %t\n", router.IsRunning())

	// Close router
	router.Close()

	fmt.Printf("router.IsRunning after close: %t\n", router.IsRunning())

	select {
	case <-router.Running():
		fmt.Println("Running")
	default:
		fmt.Println("Not running")
	}

	if err := router.Run(ctx); err != nil {
		panic(err)
	}
}

func printMessages(msg *message.Message) error {
	fmt.Printf(
		"\n> Received message: %s\n> %s\n> metadata: %v\n\n",
		msg.UUID, string(msg.Payload), msg.Metadata,
	)
	return nil
}

type structHandler struct {
	// we can add some dependencies here
}

func (s structHandler) Handler(msg *message.Message) ([]*message.Message, error) {
	log.Println("structHandler received message", msg.UUID)

	msg = message.NewMessage(watermill.NewUUID(), []byte("message produced by structHandler"))
	return message.Messages{msg}, nil
}
