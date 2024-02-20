package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type Link struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"link"`
	Duration    string `json:"duration,omitempty"`
}

type Config struct {
	Links           []Link `json:"links"`
	Duration        string `json:"duration"`
	DefaultDuration time.Duration
}

type PingData struct {
	ID           int       `json:"id"`
	IsUp         bool      `json:"is_up"`
	ResponseTime int64     `json:"response_time"`
	StatusCode   int       `json:"status_code"`
	Time         time.Time `json:"time_pinged"`
	Description  string    `json:"description"`
}

var pingData map[int]PingData

func init() {
	pingData = make(map[int]PingData)
}

func main() {
	router := gin.Default()

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

	router.GET("/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, pingData)
	})

	router.Run(":8080")
}

func loadConfig(filename string, config *Config) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return err
	}

	return nil
}

func pingLinks(links []Link, defaultDuration time.Duration) {
	for _, link := range links {
		duration := defaultDuration
		if link.Duration != "" {
			d, err := time.ParseDuration(link.Duration)
			if err != nil {
				log.Printf("Error parsing duration for link %s: %s", link.Title, err)
				continue
			}
			duration = d
		}
		go func(link Link, duration time.Duration) {
			for {
				pingLink(link)
				time.Sleep(duration)
			}
		}(link, duration)
	}
}

func pingLink(link Link) {
	start := time.Now()
	resp, err := http.Get(link.URL)
	responseTime := time.Since(start).Milliseconds()

	var isUp bool
	var statusCode int

	if err != nil {
		isUp = false
		statusCode = http.StatusInternalServerError
	} else {
		isUp = true
		statusCode = resp.StatusCode
	}

	pingData[link.ID] = PingData{
		ID:           link.ID,
		IsUp:         isUp,
		ResponseTime: responseTime,
		StatusCode:   statusCode,
		Time:         time.Now(),
		Description:  link.Description,
	}
}
