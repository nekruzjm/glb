package balancer

import (
	"math/rand"
	"net/url"
	"time"
)

func (b *balancer) Random() *url.URL {
	b.mu.RLock()
	defer b.mu.RUnlock()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return b.backends[r.Intn(len(b.backends))].url
}
