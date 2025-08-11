package database

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	Channel string `json:"channel"`
	Payload []byte `json:"payload"`
}

// Listener interface connects to the database and listen to a channel
// WaitForNotification blocks until receiving a notification or
// until the supplied context expires.
type Listener interface {
	Close(ctx context.Context) error
	Connect(ctx context.Context) error
	ListenToChannel(ctx context.Context, channel string) error
	Ping(ctx context.Context) error
	UnlistenToChannel(ctx context.Context, channel string) error
	WaitForNotification(ctx context.Context) error
}

func NewListener(dbPool *pgxpool.Pool) Listener {
	return &listener{
		mu:     sync.Mutex{},
		dbPool: dbPool,
	}
}

type listener struct {
	conn   *pgxpool.Conn
	dbPool *pgxpool.Pool
	mu     sync.Mutex
}

func (listener *listener) Close(ctx context.Context) error {
	listener.mu.Lock()
	defer listener.mu.Unlock()

	if listener.conn == nil {
		return nil
	}

	// Release below would take care of cleanup and potentially put the
	// connection back into rotation, but in case a Listen was invoked without a
	// subsequent Unlisten on the same topic, close the connection explicitly to
	// guarantee no other caller will receive a partially tainted connection.
	err := listener.conn.Conn().Close(ctx)

	// Even in the event of an error, make sure conn is set back to nil so that
	// the listener can be reused.
	listener.conn.Release()
	listener.conn = nil

	return err
}

// Connect to the database.
func (listener *listener) Connect(ctx context.Context) error {
	listener.mu.Lock()
	defer listener.mu.Unlock()

	if listener.conn != nil {
		return errors.New("connection already established")
	}

	conn, err := listener.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}

	listener.conn = conn
	return nil
}

// Listen sends a LISTEN command to a given channel
func (listener *listener) ListenToChannel(ctx context.Context, channel string) error {
	listener.mu.Lock()
	defer listener.mu.Unlock()

	_, err := listener.conn.Exec(ctx, "LISTEN \""+channel+"\"")
	return err
}

// Ping the database
func (listener *listener) Ping(ctx context.Context) error {
	listener.mu.Lock()
	defer listener.mu.Unlock()

	return listener.conn.Ping(ctx)
}

// Unlisten sends a UNLISTEN command to a given channel
func (listener *listener) UnlistenToChannel(ctx context.Context, channel string) error {
	listener.mu.Lock()
	defer listener.mu.Unlock()

	_, err := listener.conn.Exec(ctx, "UNLISTEN \""+channel+"\"")
	return err
}

func (listener *listener) WaitForNotification(ctx context.Context) error {
	listener.mu.Lock()
	defer listener.mu.Unlock()

	pgNotification, err := listener.conn.Conn().WaitForNotification(ctx)

	if err != nil {
		return err
	}

	notification := Notification{
		Channel: pgNotification.Channel,
		Payload: []byte(pgNotification.Payload),
	}

	fmt.Println(notification)

	return nil
}
