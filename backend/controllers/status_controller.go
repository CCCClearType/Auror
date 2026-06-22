package controllers

import (
	"auror_vapor_backend/database"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CheckIlearnStatus returns the most recent cached ping result from the background job.
// This avoids live-pinging iLearn on every user request, preventing rate-limit issues.
func CheckIlearnStatus(c *gin.Context) {
	type LatestPing struct {
		LatencyMs int    `json:"latency_ms"`
		Status    string `json:"status"`
	}
	var latest LatestPing
	result := database.DB.Raw("SELECT latency_ms, status FROM ilearn_pings ORDER BY checked_at DESC LIMIT 1").Scan(&latest)

	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":     "UNKNOWN",
			"error":      "No ping data yet — background worker may still be starting up",
			"latency_ms": 0,
		})
		return
	}

	resp := gin.H{
		"status":     latest.Status,
		"latency_ms": latest.LatencyMs,
	}
	if latest.Status == "DOWN" {
		resp["error"] = "All ping attempts failed"
	}
	c.JSON(http.StatusOK, resp)
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

