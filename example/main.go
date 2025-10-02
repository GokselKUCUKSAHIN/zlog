package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/GokselKUCUKSAHIN/zlog"
)

func processOrder(orderID string) error {
	return errors.New("simulated database connection timeout")
}

func processPayment() {
	zlog.Info().Segment("payment", "process").Message("Payment processing initiated")
}

func main() {
	println("=== BEFORE CONFIG ===")
	zlog.Error().Err(errors.New("test error")).Msg("No source or callstack for this error")
	zlog.Warn().Message("No source info for this warning")
	zlog.Debug().Message("No source or callstack for this debug")
	zlog.Info().Message("No source info for this info")

	println("\n=== AFTER CONFIG ===")
	zlog.SetConfig(zlog.Configure(
		zlog.AutoSourceConfig(slog.LevelError, true),
		zlog.AutoCallStackConfig(slog.LevelError, true),
		zlog.MaxCallStackDepthConfig(slog.LevelError, 8),
		zlog.AutoSourceConfig(slog.LevelWarn, true),
		zlog.AutoSourceConfig(slog.LevelInfo, true),
		zlog.AutoSourceConfig(slog.LevelDebug, true),
		zlog.AutoCallStackConfig(slog.LevelDebug, true),
		zlog.MaxCallStackDepthConfig(slog.LevelDebug, 12),
	))

	zlog.Error().Err(errors.New("test error")).Msg("Now error has automatic source and callstack")
	zlog.Warn().Message("Now warning has automatic source info")
	zlog.Debug().Message("Now debug has automatic source and callstack")
	zlog.Info().Message("Now info has automatic source info")

	println("\n=== REALISTIC SCENARIO ===")
	ctx := context.WithValue(context.Background(), "userID", "12345")
	ctx = context.WithValue(ctx, "requestID", "req-abc-123")

	if err := processOrder("order-456"); err != nil {
		zlog.Error().
			Context(ctx, []string{"userID", "requestID"}).
			Segment("order", "process").
			Err(err).
			Msgf("taskId: %s", "task-789")
	}

	processPayment()

	zlog.Info().
		Context(ctx, []string{"userID", "requestID"}).
		Segment("user", "profile", "update").
		Message("User profile updated successfully")
}
