package main

import (
	"runtime"
	"sync"
	"time"
	
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const maxPingRecords = 100

var pingData map[int]PingData
var (
	memUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mem_usage",
		Help: "Current memory usage in bytes.",
	})
	peakMemUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "peak_mem_usage",
		Help: "Peak memory usage in bytes.",
	})
	nextGC = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "next_gc_time",
		Help: "Next garbage collection (GC) time in Unix timestamp format.",
	})
	mtx sync.Mutex
)

func init() {
	pingData = make(map[int]PingData)
	prometheus.MustRegister(memUsage)
	prometheus.MustRegister(peakMemUsage)
	prometheus.MustRegister(nextGC)
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())
	api := router.Group("/api")

	var config Config
	if err := loadConfig("config.json", &config); err != nil {
		panic(err)
	}

	defaultDuration, err := time.ParseDuration(config.Duration)
	if err != nil {
		panic(err)
	}
	config.DefaultDuration = defaultDuration

	pingLinks(config.Links, config.DefaultDuration)

	ticker := time.NewTicker(defaultDuration)
	go func() {
		for {
			<-ticker.C
			pingLinks(config.Links, defaultDuration)
		}
	}()

	api.GET("/records", getRecordsHandler)

	api.GET("/records/:id", getRecordByIDHandler)

	go func() {
		for range time.Tick(5 * time.Second) {
			mtx.Lock()
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			memUsage.Set(float64(m.HeapAlloc))
			peakMemUsage.Set(float64(m.HeapSys))
			nextGC.Set(float64(m.NextGC) / 1e9)
			mtx.Unlock()
		}
	}()

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	address := config.Address + ":" + config.Port
	router.Run(address)
}
