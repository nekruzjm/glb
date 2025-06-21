package heartbeat

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/nekruzjm/glb/pkg/config"
	"github.com/nekruzjm/glb/pkg/logger"
)

type Checker interface {
	Run(doneCh <-chan struct{}, backends []string) error
	Stop(doneCh chan<- struct{})
}

type heartbeat struct {
	client *http.Client
	log    logger.Logger
}

func New(cfg config.Config, log logger.Logger) Checker {
	return &heartbeat{
		log: log,
		client: &http.Client{
			Timeout: cfg.GetDuration("heartbeat.clientTimeout") * time.Second,
		},
	}
}

var ErrNoBackends = errors.New("heartbeat error: no backends provided")

func (h *heartbeat) Run(doneCh <-chan struct{}, backends []string) error {
	if len(backends) == 0 {
		return ErrNoBackends
	}

	type res struct {
		backend    string
		err        error
		statusCode int
	}

	var (
		wg    = new(sync.WaitGroup)
		resCh = make(chan res)
	)

	wg.Add(len(backends))

	for _, backend := range backends {
		go func() {
			defer wg.Done()

			select {
			case <-doneCh:
				h.log.Warning("Heartbeat check cancelled", zap.String("backend", backend))
				return
			default:
				resp, err := h.client.Get(backend)
				if err != nil {
					resCh <- res{
						backend:    backend,
						err:        errors.New("heartbeat error: " + err.Error()),
						statusCode: http.StatusInternalServerError,
					}
					return
				}
				resCh <- res{
					backend:    backend,
					err:        nil,
					statusCode: resp.StatusCode,
				}
				_ = resp.Body.Close()
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resCh)
	}()

	for result := range resCh {
		if result.err != nil {
			h.log.Error("Heartbeat check failed",
				zap.Int("statusCode", result.statusCode),
				zap.String("backend", result.backend),
				zap.Error(result.err))
			continue
		}
		if result.statusCode != http.StatusOK {
			h.log.Warning("Heartbeat check returned non-OK status",
				zap.Int("statusCode", result.statusCode),
				zap.String("backend", result.backend))
		} else {
			h.log.Info("Heartbeat check successful",
				zap.Int("statusCode", result.statusCode),
				zap.String("backend", result.backend))
		}
	}

	return nil
}

func (h *heartbeat) Stop(doneCh chan<- struct{}) {
	doneCh <- struct{}{}
}
