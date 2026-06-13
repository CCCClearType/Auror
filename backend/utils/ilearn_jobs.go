package utils

import (
	"auror_vapor_backend/database"
	"net/http"
	"time"
)

// StartIlearnPingJob checks iLearn status every 30 seconds and saves it to DB
func StartIlearnPingJob() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Run once immediately
	pingIlearnAndSave()

	for {
		<-ticker.C
		pingIlearnAndSave()
	}
}

func pingIlearnAndSave() {
	targetURL := "https://ilearn.fcu.edu.tw/index.php"
	
	var totalLatency int64
	var successCount int

	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", targetURL, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		
		client := http.Client{Timeout: 3 * time.Second}

		start := time.Now()
		resp, err := client.Do(req)
		
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 400 {
				successCount++
				totalLatency += time.Since(start).Milliseconds()
			}
		}

		if i < 4 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	status := "DOWN"
	var finalLatency int64

	// If at least one ping succeeded, we consider it UP but average the latency of successful ones
	if successCount > 0 {
		status = "UP"
		finalLatency = totalLatency / int64(successCount)
	} else {
		finalLatency = 3000 // Represent the full timeout if completely down
	}

	database.DB.Exec("INSERT INTO ilearn_pings (checked_at, latency_ms, status) VALUES (?, ?, ?)", time.Now(), finalLatency, status)
}
