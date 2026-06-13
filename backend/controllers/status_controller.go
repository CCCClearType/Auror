package controllers

import (
	"auror_vapor_backend/database"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CheckIlearnStatus pings the FCU iLearn website and returns its status
func CheckIlearnStatus(c *gin.Context) {
	targetURL := "https://ilearn.fcu.edu.tw/index.php"

	req, _ := http.NewRequest("GET", targetURL, nil)
	// 加入瀏覽器的 User-Agent 避免被學校的防火牆擋下造成 EOF
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":     "DOWN",
			"error":      err.Error(),
			"latency_ms": latency,
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		c.JSON(http.StatusOK, gin.H{
			"status":     "UP",
			"latency_ms": latency,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":      "DOWN",
			"error":       "HTTP " + resp.Status,
			"latency_ms":  latency,
		})
	}
}

// SubmitIlearnReport handles a user reporting an issue
func SubmitIlearnReport(c *gin.Context) {
	database.DB.Exec("INSERT INTO ilearn_reports (reported_at) VALUES (?)", time.Now())
	c.JSON(http.StatusOK, gin.H{"message": "Report submitted"})
}

// GetIlearnHistory returns ping and report data for a specific time range
func GetIlearnHistory(c *gin.Context) {
	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		hours = 24
	}

	type PingRecord struct {
		CheckedAt time.Time `json:"checked_at"`
		LatencyMs int       `json:"latency_ms"`
		Status    string    `json:"status"`
	}
	type ReportRecord struct {
		ReportedAt time.Time `json:"reported_at"`
	}

	var pings []PingRecord
	var reports []ReportRecord

	timeAgo := time.Now().Add(-time.Duration(hours) * time.Hour)

	// Fetch pings
	database.DB.Raw("SELECT checked_at, latency_ms, status FROM ilearn_pings WHERE checked_at >= ? ORDER BY checked_at ASC", timeAgo).Scan(&pings)
	
	// Fetch reports
	database.DB.Raw("SELECT reported_at FROM ilearn_reports WHERE reported_at >= ? ORDER BY reported_at ASC", timeAgo).Scan(&reports)

	c.JSON(http.StatusOK, gin.H{
		"pings":   pings,
		"reports": reports,
	})
}

