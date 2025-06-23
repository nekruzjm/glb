package balancer

import (
	"net/url"
	"sort"
	"sync/atomic"
)

func (b *balancer) LeastConnections() *url.URL {
	sort.Slice(b.backends, func(i, j int) bool {
		return b.backends[i].reqCount < b.backends[j].reqCount
	})

	atomic.AddInt32(&b.backends[0].reqCount, 1)
	return b.backends[0].url
}
