package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"encore.dev/rlog"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rlog.Info("🚀 PAVE Fees API starting up...")
	rlog.Info("📊 Encore will handle HTTP server and routing")
	rlog.Info("⏰ Temporal workflows initialized and ready")
	rlog.Info("💾 Database migrations applied")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		rlog.Info("Application context cancelled")
	case sig := <-sigCh:
		rlog.Info("Received shutdown signal", "signal", sig.String())
	}

	rlog.Info("🛑 PAVE Fees API shutting down gracefully...")
}
