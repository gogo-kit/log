package main

import (
	"context"
	"errors"

	"gogokit/log"
)

func charge() error { return log.Wrap(errors.New("gateway timeout")) }

func main() {
	// Custom keys: error_message -> error, request_id -> trace_id.
	log.SetDefault(log.New(log.Config{
		Service:     "order-service",
		Environment: "production",
		Keys:        log.Keys{ErrorMessage: "error", RequestID: "trace_id"},
	}))

	// request_id/user_id flow through context.
	ctx := log.WithUserID(log.WithRequestID(context.Background(), "req-123"), "u-9")

	log.InfoContext(ctx, "order created", log.Event("ORDER_CREATED"))
	log.ErrorContext(ctx, charge(), "failed to create order")
}
