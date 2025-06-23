package proxy

import (
	"errors"
	"net/http"
	"net/http/httputil"

	"go.uber.org/zap"

	"github.com/nekruzjm/glb/internal/balancer"
	"github.com/nekruzjm/glb/pkg/config"
	"github.com/nekruzjm/glb/pkg/logger"
)

func New(cfg config.Config, log logger.Logger) (*http.Server, error) {
	lb, err := balancer.RegisterBackends(cfg, log)
	if err != nil {
		return nil, err
	}

	router := http.NewServeMux()

	router.HandleFunc("/", func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			backend := lb.LeastConnections()

			proxy := &httputil.ReverseProxy{
				Director: func(req *http.Request) {
					req.URL.Scheme = backend.Scheme
					req.URL.Host = backend.Host
					req.Host = backend.Host
				},
			}
			proxy.ServeHTTP(w, r)
		}
	}())

	server := &http.Server{
		Addr:    cfg.GetString("appPort"),
		Handler: router,
	}

	go func() {
		log.Info("Application started", zap.String("addr", server.Addr))
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Error starting server", zap.Error(err))
			return
		}
	}()

	return server, nil
}
