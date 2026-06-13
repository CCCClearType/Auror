package utils

import (
	"auror_vapor_backend/database"
	"log"
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
	req, _ := http.NewRequest("GET", targetURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	
	client := http.Client{Timeout: 5 * time.Second}

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()

	status := "DOWN"
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			status = "UP"
		}
	} else {
		// If error occurs, latency might be the full timeout, just record it
		log.Println("iLearn ping error:", err)
	}

	database.DB.Exec("INSERT INTO ilearn_pings (checked_at, latency_ms, status) VALUES (?, ?, ?)", time.Now(), latency, status)
}
