package carrot

import (
	"context"
	"errors"
	"fmt"

	"github.com/ar3s3ru/go-carrot/handler"
	"github.com/ar3s3ru/go-carrot/listener"
	"github.com/ar3s3ru/go-carrot/topology"

	"github.com/streadway/amqp"
)

// ErrNoConnection is returned by Runner.Run when no valid AMQP connection
// has been specified.
var ErrNoConnection = errors.New("carrot: no connection provided")

// ErrNoHandler is returned by Runner.Run when no handler has been specified,
// so that Runner.Run can't handle any incoming messages.
var ErrNoHandler = errors.New("carrot: no handler specified")

// ErrNoListener is returned by Runner.Run when no delivery listener
// has been specified, so that Runner.Run can't receive any messages
// from the AMQP broker.
var ErrNoListener = errors.New("carrot: no listener specified")

// Closer allows to close the amqp.Connection provided and
// any active Listener after Runner.Run has called.
type Closer struct {
	conn   *amqp.Connection
	closer listener.Closer
}

// Close closes both the amqp.Connection provided and the Listener
// declared in the Runner.
func (closer Closer) Close(ctx context.Context) error {
	defer closer.conn.Close()
	return closer.closer.Close(ctx)
}

// Closed returns a channel that gets closed when the Listener gets closed.
//
// Useful to wait for consumers completion.
func (closer Closer) Closed() <-chan error {
	return closer.closer.Closed()
}

// Runner instruments all the different parts of the go-carrot library,
// provided with a valid AMQP connection.
type Runner struct {
	conn     *amqp.Connection
	declarer topology.Declarer
	handler  handler.Handler
	listener listener.Listener

	shutdown         *Shutdown
	gracefulShutdown bool
}

// Run starts all the different parts of the Runner instrumentator,
// in the following order: topology declaration, delivery listener and messages listener.
//
// Message listener uses the sink channel coming from the delivery listener,
// and spawns a separate worker goroutine to run the message handler
// specified during configuration with the new amqp.Delivery received.
//
// An error is returned if the supplied parameters during configuration are not
// valid, or if something happened on the AMQP connection.
func (runner Runner) Run() (Closer, error) {
	if runner.conn == nil {
		return Closer{}, ErrNoConnection
	}

	if runner.declarer != nil {
		if err := runner.declareTopology(); err != nil {
			return Closer{}, fmt.Errorf("carrot: failed to declare topology, %w", err)
		}
	}

	// No handler nor delivery listener is an acceptable scenario: it means
	// the user is not leveraging carrot for message consumption.
	if runner.handler == nil && runner.listener == nil {
		return Closer{}, nil
	}

	closer, err := runner.listenAndServe()
	if err != nil {
		return Closer{}, fmt.Errorf("carrot: failed to listen and serve consumers, %w", err)
	}

	runnerCloser := Closer{
		conn:   runner.conn,
		closer: closer,
	}

	if runner.gracefulShutdown {
		go gracefulShutdown(runnerCloser, runner.shutdown.orDefault())
	}

	return runnerCloser, nil
}

func (runner Runner) declareTopology() error {
	ch, err := runner.openChannel()
	if err != nil {
		return err
	}

	defer ch.Close()

	return runner.declarer.Declare(ch)
}

func (runner Runner) listenAndServe() (listener.Closer, error) {
	if runner.handler == nil {
		return nil, ErrNoHandler
	}

	if runner.listener == nil {
		return nil, ErrNoListener
	}

	ch, err := runner.openChannel()
	if err != nil {
		return nil, err
	}

	closer, err := runner.listener.Listen(runner.conn, ch, runner.handler)
	if err != nil {
		return nil, fmt.Errorf("carrot: failed to listen, %w", err)
	}

	return closer, nil
}

func (runner Runner) openChannel() (*amqp.Channel, error) {
	ch, err := runner.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("carrot: failed to create channel from connection, %w", err)
	}

	return ch, nil
}

// From creates a new Runner instance, given an AMQP connection and options.
//
// Required options are WithListener, to bind a channel to an amqp.Delivery sink
// and start receiving messages, and WithHandler, to handle all the incoming
// messages.
func From(conn *amqp.Connection, options ...Option) Runner {
	runner := Runner{conn: conn}

	for _, option := range options {
		if option == nil {
			continue
		}

		option(&runner)
	}

	return runner
}

// Option represents an additional argument for the Runner factory method.
type Option func(*Runner)

// WithTopology adds a topology declaration step to the new Runner instance.
func WithTopology(declarer topology.Declarer) Option {
	return func(runner *Runner) { runner.declarer = declarer }
}

// WithHandler specifies the component in charge of handling incoming messages
// for the new Runner instance.
func WithHandler(handler handler.Handler) Option {
	return func(runner *Runner) { runner.handler = handler }
}

// WithListener specifies the component in charge of start listening messages
// coming from the AMQP broker.
func WithListener(listener listener.Listener) Option {
	return func(runner *Runner) { runner.listener = listener }
}

// WithGracefulShutdown enables graceful shutdown after certain signals
// are received by the process.
//
// Use WithGracefulShutdown(nil) to use the default options, which can be found
// in DefaultShutdownOptions.
func WithGracefulShutdown(options *Shutdown) Option {
	return func(runner *Runner) {
		runner.shutdown = options
		runner.gracefulShutdown = true
	}
}
