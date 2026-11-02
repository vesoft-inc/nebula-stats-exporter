package exporter

import (
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestCollectFromStaticConfigLimitsConcurrency(t *testing.T) {
	const instanceCount = 50
	const concurrencyLimit = 3

	var active, maxActive atomic.Int32
	started := make(chan struct{}, instanceCount)
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseAll := func() {
		releaseOnce.Do(func() {
			close(release)
		})
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := active.Add(1)
		defer active.Add(-1)

		for previous := maxActive.Load(); current > previous; previous = maxActive.Load() {
			if maxActive.CompareAndSwap(previous, current) {
				break
			}
		}

		started <- struct{}{}
		<-release
	}))
	defer server.Close()
	defer releaseAll()

	port := int32(server.Listener.Addr().(*net.TCPAddr).Port)
	instances := make([]Instance, instanceCount)
	for i := range instances {
		instances[i] = Instance{
			EndpointIP:    "127.0.0.1",
			EndpointPort:  port,
			ComponentType: ComponentTypeGraphd,
		}
	}

	exporter := &NebulaExporter{
		MaxConcurrentMetricRequests: concurrencyLimit,
		Config: StaticConfig{
			Clusters: []Cluster{{Instances: instances}},
		},
	}
	done := make(chan struct{})
	go func() {
		exporter.CollectFromStaticConfig(make(chan prometheus.Metric, instanceCount))
		close(done)
	}()

	for i := range concurrencyLimit {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			t.Fatalf("only %d metric requests started; want %d", i, concurrencyLimit)
		}
	}

	select {
	case <-done:
		t.Fatal("collection completed before blocked scrape requests were released")
	default:
	}

	select {
	case <-started:
		t.Fatal("more than the configured number of metric requests started concurrently")
	case <-time.After(100 * time.Millisecond):
	}

	releaseAll()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("collection did not finish after scrape requests were released")
	}

	if got := maxActive.Load(); got > concurrencyLimit {
		t.Errorf("maximum concurrent metric requests = %d, want at most %d", got, concurrencyLimit)
	}
}
