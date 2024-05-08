package main

import (
	"log"
	"net/http"
	"time"
)



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

    if existingData, ok := pingData[link.ID]; ok {
        if len(existingData.PingTimes) >= maxPingRecords {
            existingData.PingTimes = existingData.PingTimes[1:]
        }
        existingData.PingTimes = append(existingData.PingTimes, PingTime{
            Time:         time.Now(),
            ResponseTime: responseTime,
        })
        existingData.IsUp = isUp
        existingData.StatusCode = statusCode
        existingData.Time = time.Now()
        existingData.URL = link.URL
        pingData[link.ID] = existingData
    } else {
        pingData[link.ID] = PingData{
            ID:          link.ID,
            IsUp:        isUp,
            PingTimes:   []PingTime{{Time: time.Now(), ResponseTime: responseTime}},
            StatusCode:  statusCode,
            Time:        time.Now(),
            Description: link.Description,
            URL:         link.URL,
        }
    }
}



/*

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

	if existingData, ok := pingData[link.ID]; ok {
		if len(existingData.PingTimes) >= maxPingRecords {
			existingData.PingTimes = existingData.PingTimes[1:]
		}
		existingData.PingTimes = append(existingData.PingTimes, PingTime{
			Time:         time.Now(),
			ResponseTime: responseTime,
		})
		existingData.IsUp = isUp
		existingData.StatusCode = statusCode
		existingData.Time = time.Now()
		pingData[link.ID] = existingData
	} else {
		pingData[link.ID] = PingData{
			ID:          link.ID,
			IsUp:        isUp,
			PingTimes:   []PingTime{{Time: time.Now(), ResponseTime: responseTime}},
			StatusCode:  statusCode,
			Time:        time.Now(),
			Description: link.Description,
		}
	}
}
*/





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
