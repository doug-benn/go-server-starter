package repository

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/doug-benn/go-server-starter/database"
	"github.com/doug-benn/go-server-starter/producer"
	"github.com/doug-benn/go-server-starter/sse"
)

const eventChannelBuffer = 100

type DatabaseEvent struct {
	Table     string         `json:"table"`
	Action    string         `json:"action"`
	Timestamp time.Time      `json:"timestamp"`
	Data      map[string]any `json:"record"`
}

func DecodeAsDatabaseEvent(payload []byte) (*DatabaseEvent, error) {
	var event DatabaseEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

func NotificationProcessing(ctx context.Context, logger *slog.Logger, postgresListener database.Listener, sseProducer *producer.Producer[sse.Event]) {
	eventCh := make(chan sse.Event, eventChannelBuffer)

	go drainAndBroadcast(ctx, eventCh, sseProducer)

	for {
		notification, err := postgresListener.WaitForNotification(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			logger.Error("error waiting for notification", "error", err)
			continue
		}

		payload, err := DecodeAsDatabaseEvent(notification.Payload)
		if err != nil {
			logger.Error("decode error", "error", err)
			continue
		}

		logger.Info("database event received",
			"table", payload.Table,
			"action", payload.Action,
		)

		select {
		case eventCh <- sse.Event{Data: payload}:
		default:
			logger.Warn("drain too slow, dropping notification",
				"table", payload.Table,
				"action", payload.Action,
				"channel_capacity", cap(eventCh),
				"channel_usage", len(eventCh),
			)
		}
	}
}

func drainAndBroadcast(ctx context.Context, eventCh <-chan sse.Event, sseProducer *producer.Producer[sse.Event]) {
	for {
		select {
		case event := <-eventCh:
			sseProducer.Broadcast(ctx, event)
		case <-ctx.Done():
			return
		}
	}
}
