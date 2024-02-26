package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/redis/go-redis/v9"
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

var (
	ctx = context.Background()

	rdb *redis.Client
)

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}

func main() {
	router := gin.Default()
	router.GET("/data", func(c *gin.Context) {
		pingData, err := rdb.HGetAll(context.Background(), "ping_data").Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve ping data"})
			return
		}

		var pingDataSlice []PingData
		for _, v := range pingData {
			var pd PingData
			err := json.Unmarshal([]byte(v), &pd)
			if err != nil {
				log.Printf("Error unmarshaling ping data: %s", err)
				continue
			}
			pingDataSlice = append(pingDataSlice, pd)
		}

		c.JSON(http.StatusOK, pingDataSlice)
	})

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

	data := map[string]interface{}{
		"id":            link.ID,
		"is_up":         isUp,
		"response_time": responseTime,
		"status_code":   statusCode,
		"time_pinged":   time.Now().Format(time.RFC3339),
		"description":   link.Description,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling ping data for link %s: %s", link.Title, err)
		return
	}

	err = rdb.HSet(ctx, "ping_data", strconv.Itoa(link.ID), jsonData).Err()
	if err != nil {
		log.Printf("Error storing ping data for link %s in Redis: %s", link.Title, err)
	}
}
