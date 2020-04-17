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

### tl;dr features list

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

## Architecture

Carrot uses _three main components_ for its API:

1. [**Topology declarator**](#topology-declarator), to declare the AMQP topology from the application that uses it
2. [**Message handlers**](#message-handlers), to define functions able to handle incoming messages from consumers
3. [**Consumer listeners**](#consumer-listeners), to receive messages from one or more queues


### Topology declarator

Carrot allows to define a topology by exposing an expressive API backed by
[`topology.Declarer`](topology/topology.go#L8-L10) interface.

The current supported topologies are:
* **Queues**, found in [`topology/queue` package](topology/queue/queue.go)
* **Exchanges**, found in [`topology/exchange` package](topology/exchange/exchange.go)

Topology declaration is **optional**, and can be controlled with `carrot.WithTopology`:

```go
carrot.WithTopology(topology.All(
    exchange.Declare("messages"),
    queue.Declare("consumer.message.received",
        queue.BindTo("messages", "message.published"),
    ),
    queue.Declare("consumer.message.deleted",
        queue.BindTo("messages", "message.deleted"),
    ),
)),
```

When specified, Carrot will open a dedicated AMQP channel
to declare the topology, before listening to messages.

Carrot can also be used **exclusively** for topology declaration:

```go
conn, err := amqp.Dial("amqp://guest:guest@rabbit:5672")
if err != nil {
    panic(err)
}

// The carrot.Closer handle returned is useless if only topology declaration
// is used.
_, err := carrot.Run(conn,
    // Declare your topology here
    carrot.WithTopology(topology.All(
        // topology.All is used to declare more than one topology
        // in a single transaction.
        exchange.Declare("messages"),
        queue.Declare("consumer.message.received",
            queue.BindTo("messages", "message.published"),
        ),
        queue.Declare("consumer.message.deleted",
            queue.BindTo("messages", "message.deleted"),
        ),
    )),
)
```

### Message handlers

Carrot defines an interface for handling incoming messages (`amqp.Delivery`)
in [`handler.Handler` interface](handler/handler.go):

```go
type Handler interface {
    Handle(context.Context, amqp.Delivery) error
}
```

Message handlers are **fallible**, so they can return an error.

Error handling can be specified at [Consumer Listeners](#consumer-listeners) level.

You can specify a message handler for all incoming messages by using `carrot.WithHandler`:

```go
carrot.WithHandler(handler.Func(func(context.Context, amqp.Delivery) error {
    // Handle messages here!
    return nil
}))
```

#### Router

Carrot also exposes a [`Router`](handler/router/router.go) interface and implementation
to support:

* Multiple listeners with their own message handlers
* Middleware support

An example of how a `Router` setup might look like:

```go
// Router implements the handler.Handler interface.
router.New().Group(func(r router.Router) {
    // This is how you set middlewares.
    r.Use(LogMessages(logger))
    r.Use(middleware.Timeout(50 * time.Millisecond))
    r.Use(SimulateWork(100*time.Millisecond, logger))

    // This is how you bind an handler function to a specific queue.
    // In order for it to work, you must register these queues
    // in the listener.
    r.Bind("consumer.message.received", handler.Func(Acknowledger))
    r.Bind("consumer.message.deleted", handler.Func(Acknowledger))

    // You can also specify additional middlewares only for one queue:
    r.With(AuthenticateUser).
        Bind("consumer.message.created", handler.Func(Acknowledger))
})
```

### Consumer listeners

As the name says, Listeners listens for incoming messages on a specific queue.

Carrot defines a [`listener.Listener` interface](listener/listener.go) to
represent these components:

```go
type Listener interface {
    Listen(Connection, Channel, handler.Handler) (Closer, error)
}
```

so that the listener can:

- Start **listening** to incoming `amqp.Delivery` from a `Channel`
- **Serving** these messages using the provided `handler.Handler`
- Hand out a [`Closer` handler](listener/listener.go) to **close** the listener/server goroutine and/or **wait for its closing**

An example of how to define Listeners:

```go
// WithListener specifies the listener.Listener to start.
carrot.WithListener(listener.Sink(
    // listener.Sink allows to listen to messages coming from one or more consumers,
    // and pilots closing the child listeners.
    consumer.Listen("consumer.message.deleted"),
    listener.UseDedicatedChannel(
        // By default, carrot uses a single amqp.Channel to establish
        // consumer listeners. But we can tell carrot to use a dedicated
        // amqp.Channel for certain consumers.
        consumer.Listen("consumer.message.received"),
    ),
))
```

## Full example

Let's put all the pieces together now!

```go
conn, err := amqp.Dial("amqp://guest:guest@rabbit:5672")
if err != nil {
    panic(err)
}

closer, err := carrot.Run(conn,
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
        consumer.Listen("consumer.message.deleted"),
        listener.UseDedicatedChannel(
            consumer.Listen("consumer.message.received"),
        ),
    ))
    // Lastly, specify an handler function that will receive the messages
    // coming from the specified consumers.
    carrot.WithHandler(router.New().Group(func(r router.Router) {
        r.Use(LogMessages(logger))
        r.Use(middleware.Timeout(50 * time.Millisecond))
        r.Use(SimulateWork(100*time.Millisecond, logger))

        r.Bind("consumer.message.received", handler.Func(Acknowledger))
        r.Bind("consumer.message.deleted", handler.Func(Acknowledger))
    })),
)

if err != nil {
    panic(err)
}

// Wait on the main goroutine until the consumer has exited:
err := <-closer.Closed()
log.Fatalf("Consumers closed (error %s)", err)
```

## License

This project is licensed under the [MIT license](LICENSE).

### Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted for inclusion in `go-carrot` by you, shall be licensed as MIT, without any additional terms or conditions.
