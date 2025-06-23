package proxy

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"

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

	backend := lb.Random()

	router := http.NewServeMux()
	router.HandleFunc("/", handler(httputil.NewSingleHostReverseProxy(backend), backend))

	var server = &http.Server{
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

	return server, nil
}

func handler(rp *httputil.ReverseProxy, backend *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Host = backend.Host
		r.URL.Scheme = backend.Scheme
		rp.ServeHTTP(w, r)
	}
}
