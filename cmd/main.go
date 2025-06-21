package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/nekruzjm/glb/internal/balancer"
	"github.com/nekruzjm/glb/pkg/config"
	"github.com/nekruzjm/glb/pkg/logger"
)

func main() {
	cfg := config.New()
	log := logger.New(cfg)

	backends := cfg.GetStringSlice("backends")
	lb, err := balancer.New(backends)
	if err != nil {
		if errors.Is(err, balancer.ErrEmptyBackends) {
			log.Warning("empty backends", zap.Error(err))
			return
		}
		log.Error("err occurred", zap.Error(err), zap.Strings("backends", backends))
		return
	}

	selectedBackend := lb.Random()

	handle := func(rp *httputil.ReverseProxy) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r.Host = selectedBackend.Host
			r.URL.Scheme = selectedBackend.Scheme
			rp.ServeHTTP(w, r)
		}
	}

	router := http.NewServeMux()
	router.HandleFunc("/", handle(httputil.NewSingleHostReverseProxy(selectedBackend)))

	server := &http.Server{
		Addr:    ":" + cfg.GetString("port"),
		Handler: router,
	}

	go func() {
		log.Info("Application started", zap.String("addr", server.Addr))
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Error starting server", zap.Error(err))
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop

	log.Info("Received signal", zap.String("signal", sign.String()))

	log.Flush()
	_ = server.Shutdown(context.Background())

	log.Info("Application stopped")
}
