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
	zlog.Info().Segment("payment", "process").Message("Payment işlemi başlatıldı")
}

func main() {
	// Önce config set etmeden test edelim
	println("=== CONFIG SET ETMEDEN ===")
	zlog.Error().Err(errors.New("test error")).Msg("Bu error için source ve callstack yok")
	zlog.Warn().Message("Bu warn için source bilgisi yok")
	zlog.Debug().Message("Bu debug için source ve callstack yok")
	zlog.Info().Message("Bu info için source bilgisi yok")

	println("\n=== CONFIG SET ETTIKTEN SONRA ===")
	// Global config'i set et - Yeni temiz yapı!

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

	// Şimdi aynı testleri tekrar çalıştır
	zlog.Error().Err(errors.New("test error")).Msg("Şimdi error için otomatik source ve callstack var")
	zlog.Warn().Message("Şimdi warn için otomatik source bilgisi var")
	zlog.Debug().Message("Şimdi debug için otomatik source ve callstack var")
	zlog.Info().Message("Şimdi info için otomatik source bilgisi var")

	println("\n=== GERÇEKÇİ SENARYO ===")
	// Gerçekçi bir senaryo test et
	ctx := context.WithValue(context.Background(), "userID", "12345")
	ctx = context.WithValue(ctx, "requestID", "req-abc-123")

	// Kullanıcının orijinal verbose kodu:
	// zlog.Error().WithSource().WithCallStack().Segment(command.Name.String()).Error(err).Messagef("taskId: %s", command.Task.Id)

	// Şimdi sadeleşmiş versiyonu:
	if err := processOrder("order-456"); err != nil {
		zlog.Error().
			Context(ctx, []string{"userID", "requestID"}).
			Segment("order", "process").
			Err(err).
			Msgf("taskId: %s", "task-789")
	}

	// Diğer fonksiyonlardan log çıktıları
	processPayment()

	// Hata olmayan durumlar
	zlog.Info().
		Context(ctx, []string{"userID", "requestID"}).
		Segment("user", "profile", "update").
		Message("Kullanıcı profili başarıyla güncellendi")
}
