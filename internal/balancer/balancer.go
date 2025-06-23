package balancer

import (
	"errors"
	"net/url"
	"sync"

	"go.uber.org/zap"

	"github.com/nekruzjm/glb/pkg/config"
	"github.com/nekruzjm/glb/pkg/logger"
)

type Balancer interface {
	Random() *url.URL
	LeastConnections() *url.URL
}

type balancer struct {
	backends []backend
	mu       sync.RWMutex
}

type backend struct {
	url      *url.URL
	reqCount int32
	weight   int
}

func RegisterBackends(cfg config.Config, log logger.Logger) (Balancer, error) {
	backends := cfg.GetStringSlice("backends")

	if len(backends) == 0 {
		log.Warning("balancer err: no backends provided")
		return nil, errors.New("balancer err: empty backends")
	}

	var urls = make([]backend, 0, len(backends))
	for _, b := range backends {
		parsedUrl, err := url.Parse(b)
		if err != nil {
			log.Error("balancer err: failed to parse backend", zap.String("backend", b), zap.Error(err))
			return nil, err
		}

		urls = append(urls, backend{
			url: parsedUrl,
		})
	}

	return &balancer{
		backends: urls,
	}, nil
}
