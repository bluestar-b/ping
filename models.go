package main

import "time"

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

	Port    string `json:"port"`
	Address string `json:"address"`
}

type PingTime struct {
	Time         time.Time `json:"time_pinged"`
	ResponseTime int64     `json:"response_time"`
}

type PingData struct {
	ID          int        `json:"id"`
	IsUp        bool       `json:"is_up"`
	PingTimes   []PingTime `json:"ping_times"`
	StatusCode  int        `json:"status_code"`
	Time        time.Time  `json:"last_pinged"`
	Description string     `json:"description"`
}
