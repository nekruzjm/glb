package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/nekruzjm/glb/internal/heartbeat"
	"github.com/nekruzjm/glb/internal/proxy"
	"github.com/nekruzjm/glb/pkg/config"
	"github.com/nekruzjm/glb/pkg/logger"
)

func main() {
	cfg := config.New()
	log := logger.New(cfg)

	server, err := proxy.New(cfg, log)
	if err != nil {
		return
	}

	doneCh := make(chan struct{})
	hb := heartbeat.New(cfg, log)

	go run(cfg, log, hb, doneCh)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop

	log.Info("Received signal", zap.String("signal", sign.String()))

	log.Flush()
	_ = server.Shutdown(context.Background())
	hb.Stop(doneCh)

	log.Info("Application stopped")
}

func run(cfg config.Config, log logger.Logger, hb heartbeat.HealthChecker, done <-chan struct{}) {
	var (
		hbBackends = cfg.GetStringSlice("heartbeat.backends")
		hbInterval = cfg.GetDuration("heartbeat.interval")
	)

	ticker := time.NewTicker(hbInterval * time.Second)

	for {
		select {
		case <-ticker.C:
			err := hb.Run(done, hbBackends)
			if err != nil {
				if errors.Is(err, heartbeat.ErrNoBackends) {
					log.Warning("empty backends", zap.Error(err))
				} else {
					log.Error("err occurred", zap.Error(err), zap.Strings("backends", hbBackends))
				}
			}
		case <-done:
			log.Info("Heartbeat stopped")
			return
		}
	}
}
