package metric

import (
	"errors"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/nekruzjm/glb/pkg/config"
	"github.com/nekruzjm/glb/pkg/logger"
)

func New(cfg config.Config, log logger.Logger) *http.Server {
	router := http.NewServeMux()
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewBuildInfoCollector())

	router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	metricServer := &http.Server{
		Addr:    cfg.GetString("metricPort"),
		Handler: router,
	}

	go func() {
		log.Info("Metric started", zap.String("addr", metricServer.Addr))
		if err := metricServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Error starting server", zap.Error(err))
			return
		}
	}()

	return metricServer
}
