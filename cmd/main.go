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
	"github.com/nekruzjm/glb/internal/metric"
	"github.com/nekruzjm/glb/internal/proxy"
	"github.com/nekruzjm/glb/pkg/config"
	"github.com/nekruzjm/glb/pkg/logger"
)

func main() {
	cfg := config.New()
	log := logger.New(cfg)

	proxyServer, err := proxy.New(cfg, log)
	if err != nil {
		return
	}

	metricServer := metric.New(cfg, log)

	done := make(chan struct{})
	hb := heartbeat.New(cfg, log)

	go runHeartBeat(cfg, log, hb, done)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop

	log.Info("Received signal", zap.String("signal", sign.String()))

	_ = proxyServer.Shutdown(context.Background())
	_ = metricServer.Shutdown(context.Background())
	hb.Stop(done)
	log.Flush()

	log.Info("Application stopped")
	log.Info("Metrics stopped")
}

func runHeartBeat(cfg config.Config, log logger.Logger, hb heartbeat.HealthChecker, done <-chan struct{}) {
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
