# go-carrot

<div align="center">
    <strong>
        Declarative API for AMQP consumers in Go.
    </strong>
</div>

<br />

<div align="center">
    <!-- Godoc -->
    <a href="https://godoc.org/github.com/ar3s3ru/go-carrot">
        <img alt="Godoc reference"
            src="https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square">
    </a>
    <!-- Workflow -->
    <a href="">
        <img alt="GitHub Workflow Status"
            src="https://img.shields.io/github/workflow/status/ar3s3ru/go-carrot/Pipeline?style=flat-square">
    </a>
    <!-- Coverage -->
    <a href="https://codecov.io/gh/ar3s3ru/go-carrot">
        <img alt="Codecov"
            src="https://img.shields.io/codecov/c/github/ar3s3ru/go-carrot?style=flat-square">
    </a>
    <!-- License -->
    <a href="LICENSE">
        <img alt="License"
            src="https://img.shields.io/github/license/ar3s3ru/go-carrot?style=flat-square">
    </a>
</div>

<br />

<div align="center">
    <img height=360 src="assets/logo.jpg">
</div>

<div align="center">
    <i>
        (It's supposed to be a cute logo...)
    </i>
</div>

## tl;dr features list

- [x] Topology declaration API
- [x] Consumer listener API
- [x] Message handler API
- [x] Consumers router
  - [x] Middleware support
  - [ ] Common middlewares implementation
- [x] Graceful shutdown
- [ ] Automatic reconnection

## Description

Carrot exposes a nice API for dealing with AMQP connections, such as _declaring topologies_ (exchanges, queues, ...) and _declaring consumers_ on one or more queues.

Check out the [examples](examples/simple-consumer/main.go) for more information.

### Architecture

Carrot uses _three main components_ for its API:

1. **Topology declarator**, to declare the AMQP topology from the application that uses it
2. **Consumer listeners**, to receive messages from one or more queues
3. **Message handlers**, to define functions able to handle incoming messages from consumers

### Example

```go
conn, err := amqp.Dial("amqp://guest:guest@rabbit:5672")
if err != nil {
    panic(err)
}

carrot.From(conn,
    // First, declare your topology...
    carrot.WithTopology(topology.All(
        exchange.Declare("messages"),
        queue.Declare("consumer.message.received",
            queue.BindTo("messages", "message.published"),
        ),
        queue.Declare("consumer.message.deleted",
            queue.BindTo("messages", "message.deleted"),
        ),
    )),
    // Second, declare the consumers to receive messages from...
    carrot.WithListener(listener.Sink(
        // listener.Sink allows to listen to messages coming from one or more consumers!
        listener.UseDedicatedChannel(consumer.Listen(
            // By default, carrot uses a single amqp.Channel to establish
            // consumer listeners. But we can tell carrot to use a dedicated
            // amqp.Channel for certain consumers.
            "consumer.message.received",
            consumer.Title("Message received"),
        )),
        consumer.Listen(
            "consumer.message.deleted",
            consumer.Title("Message deleted"),
        ),
    )),
    // Lastly, specify an handler function that will receive the messages
    // coming from the specified consumers.
    carrot.WithHandler(router.New().Group(func(r router.Router) {
        // Carrot provides a Multiplexer that supports multiple handlers
        // for multiple queues and middlewares.

        // This is how you set middlewares.
        r.Use(LogMessages(logger))
        r.Use(middleware.Timeout(50 * time.Millisecond))
        r.Use(SimulateWork(100*time.Millisecond, logger))

        // This is how you bind an handler function to a specific queue.
        // In order for it to work, you must register these queues
        // in the listener.
        r.Bind("consumer.message.received", handler.Func(Acknowledger))
        r.Bind("consumer.message.deleted", handler.Func(Acknowledger))
    })),
).
// ...and start listening to messages!
Run()
```

## License

This project is licensed under the [MIT license](LICENSE).

### Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted for inclusion in `go-carrot` by you, shall be licensed as MIT, without any additional terms or conditions.
