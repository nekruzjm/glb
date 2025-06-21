package balancer

import (
	"errors"
	"math/rand"
	"net/url"
	"sort"
	"sync"
	"time"
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
	reqCount int
}

var ErrEmptyBackends = errors.New("balancer err: empty backends")

func New(backends []string) (Balancer, error) {
	if len(backends) == 0 {
		return nil, ErrEmptyBackends
	}

	var urls = make([]backend, 0, len(backends))
	for _, b := range backends {
		parsedUrl, err := url.Parse(b)
		if err != nil {
			return nil, errors.New("balancer err: " + err.Error())
		}

		urls = append(urls, backend{
			url: parsedUrl,
		})
	}

	return &balancer{
		backends: urls,
	}, nil
}

func (b *balancer) Random() *url.URL {
	b.mu.RLock()
	defer b.mu.RUnlock()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return b.backends[r.Intn(len(b.backends))].url
}

func (b *balancer) LeastConnections() *url.URL {
	b.mu.Lock()
	defer b.mu.Unlock()

	sort.Slice(b.backends, func(i, j int) bool {
		return b.backends[i].reqCount < b.backends[j].reqCount
	})

	b.backends[0].reqCount++
	return b.backends[0].url
}
