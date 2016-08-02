package statsdprof

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/peterbourgon/g2s"
)

const (
	ENV_STATSD          = "STATSD_HOST"
	DEFAULT_STATSD_HOST = "172.17.42.1:8125"
	COLLECT_DELAY       = 1 * time.Minute
	SERVICE             = "[STATSD_PPROF]"
)

var (
	_statter g2s.Statter
)

func init() {
	addr := DEFAULT_STATSD_HOST
	if env := os.Getenv(ENV_STATSD); env != "" {
		addr = env
	}

	s, err := g2s.Dial("udp", addr)
	if err == nil {
		_statter = s
	} else {
		_statter = g2s.Noop()
		log.Println(err)
	}

	go pprof_task()
}

// profiling task
func pprof_task() {
	for {
		<-time.After(COLLECT_DELAY)
		collect()
	}
}

// collect & publish to statsd
func collect() {
	var tag string
	hostname, err := os.Hostname()
	if err != nil {
		log.Println(SERVICE, err)
		return
	}
	tag = hostname + ".pprof"

	// collect
	memstats := &runtime.MemStats{}
	runtime.ReadMemStats(memstats)

	_statter.Gauge(1.0, tag+".NumGoroutine", fmt.Sprint(runtime.NumGoroutine()))
	_statter.Gauge(1.0, tag+".NumCgoCall", fmt.Sprint(runtime.NumCgoCall()))
	if memstats.NumGC > 0 {
		_statter.Timing(1.0, tag+".PauseTotal", time.Duration(memstats.PauseTotalNs))
		_statter.Timing(1.0, tag+".LastPause", time.Duration(memstats.PauseNs[(memstats.NumGC+255)%256]))
		_statter.Gauge(1.0, tag+".NumGC", fmt.Sprint(memstats.NumGC))
		_statter.Gauge(1.0, tag+".Alloc", fmt.Sprint(memstats.Alloc))
		_statter.Gauge(1.0, tag+".TotalAlloc", fmt.Sprint(memstats.TotalAlloc))
		_statter.Gauge(1.0, tag+".Sys", fmt.Sprint(memstats.Sys))
		_statter.Gauge(1.0, tag+".Lookups", fmt.Sprint(memstats.Lookups))
		_statter.Gauge(1.0, tag+".Mallocs", fmt.Sprint(memstats.Mallocs))
		_statter.Gauge(1.0, tag+".Frees", fmt.Sprint(memstats.Frees))
		_statter.Gauge(1.0, tag+".HeapAlloc", fmt.Sprint(memstats.HeapAlloc))
		_statter.Gauge(1.0, tag+".HeapSys", fmt.Sprint(memstats.HeapSys))
		_statter.Gauge(1.0, tag+".HeapIdle", fmt.Sprint(memstats.HeapIdle))
		_statter.Gauge(1.0, tag+".HeapInuse", fmt.Sprint(memstats.HeapInuse))
		_statter.Gauge(1.0, tag+".HeapReleased", fmt.Sprint(memstats.HeapReleased))
		_statter.Gauge(1.0, tag+".HeapObjects", fmt.Sprint(memstats.HeapObjects))
		_statter.Gauge(1.0, tag+".StackInuse", fmt.Sprint(memstats.StackInuse))
		_statter.Gauge(1.0, tag+".StackSys", fmt.Sprint(memstats.StackSys))
		_statter.Gauge(1.0, tag+".MSpanInuse", fmt.Sprint(memstats.MSpanInuse))
		_statter.Gauge(1.0, tag+".MSpanSys", fmt.Sprint(memstats.MSpanSys))
		_statter.Gauge(1.0, tag+".MCacheInuse", fmt.Sprint(memstats.MCacheInuse))
		_statter.Gauge(1.0, tag+".MCacheSys", fmt.Sprint(memstats.MCacheSys))
		_statter.Gauge(1.0, tag+".BuckHashSys", fmt.Sprint(memstats.BuckHashSys))
		_statter.Gauge(1.0, tag+".GCSys", fmt.Sprint(memstats.GCSys))
		_statter.Gauge(1.0, tag+".OtherSys", fmt.Sprint(memstats.OtherSys))
	}
}
