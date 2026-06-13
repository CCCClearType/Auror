package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"auror_vapor_backend/database"
	"auror_vapor_backend/models"
	"auror_vapor_backend/utils"

	"github.com/gin-gonic/gin"
)

func attachRatingSummary(note *models.Note) {
	var result struct {
		Rating float64
		Count  int64
	}
	database.DB.Table("reviews").
		Select("COALESCE(AVG(CASE WHEN attitude = 'POSITIVE' THEN 5.0 ELSE 1.0 END), 0) AS rating, COUNT(*) AS count").
		Where("note_id = ? AND status = ?", note.NoteID, "VISIBLE").
		Scan(&result)
	note.OverallRating = result.Rating
	note.RatingCount = result.Count
}

func refreshStoredNoteRating(noteID uint) {
	var note models.Note
	if err := database.DB.First(&note, noteID).Error; err != nil {
		return
	}
	attachRatingSummary(&note)
	database.DB.Model(&models.Note{}).Where("note_id = ?", noteID).Update("overall_rating", note.OverallRating)
}

// GetNotes handles GET /api/notes (UC-01, UC-02)
func GetNotes(c *gin.Context) {
	var notes []models.Note
	q := c.Query("q")
	tag := c.Query("tag")
	seller := c.Query("seller")
	sort := c.DefaultQuery("sort", "price_asc")

	query := database.DB.Model(&models.Note{}).Group("notes.note_id").Where("notes.status = ?", "ACTIVE")
	if q != "" {
		keyword := "%" + q + "%"
		query = query.
			Joins("LEFT JOIN note_tags q_note_tags ON q_note_tags.note_id = notes.note_id").
			Joins("LEFT JOIN tags q_tags ON q_tags.tag_id = q_note_tags.tag_id").
			Joins("LEFT JOIN users q_sellers ON q_sellers.user_id = notes.seller_id").
			Where("notes.title ILIKE ? OR notes.description ILIKE ? OR q_tags.tag_name ILIKE ? OR q_sellers.username ILIKE ?", keyword, keyword, keyword, keyword)
	}
	if tag != "" {
		query = query.
			Joins("JOIN note_tags filter_note_tags ON filter_note_tags.note_id = notes.note_id").
			Joins("JOIN tags filter_tags ON filter_tags.tag_id = filter_note_tags.tag_id").
			Where("filter_tags.tag_name ILIKE ?", tag)
	}
	if seller != "" {
		query = query.
			Joins("JOIN users filter_sellers ON filter_sellers.user_id = notes.seller_id").
			Where("filter_sellers.username ILIKE ?", seller)
	}
	if minPrice, err := strconv.ParseFloat(c.Query("min_price"), 64); err == nil {
		query = query.Where("notes.price >= ?", minPrice)
	}
	if maxPrice, err := strconv.ParseFloat(c.Query("max_price"), 64); err == nil {
		query = query.Where("notes.price <= ?", maxPrice)
	}

	if c.Query("hide_owned") == "true" {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				if claims, err := utils.ParseToken(parts[1]); err == nil {
					userID := uint(claims["user_id"].(float64))
					query = query.Where("NOT EXISTS (SELECT 1 FROM note_licenses WHERE note_licenses.note_id = notes.note_id AND note_licenses.user_id = ? AND note_licenses.status = 'ACTIVE') AND notes.seller_id != ?", userID, userID)
				}
			}
		}
	}

	switch sort {
	case "price_desc":
		query = query.Order("notes.price DESC, notes.note_id DESC")
	case "price_asc":
		fallthrough
	default:
		query = query.Order("notes.price ASC, notes.note_id DESC")
	}

	// Retrieve notes from the database with their media
	if err := query.Preload("Media").Preload("Tags").Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notes"})
		return
	}
	for i := range notes {
		attachRatingSummary(&notes[i])
		var dev models.User
		if err := database.DB.Select("username").First(&dev, notes[i].SellerID).Error; err == nil {
			notes[i].SellerName = dev.Username
		} else {
			notes[i].SellerName = "未知"
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": notes})
}

// GetNoteByID handles GET /api/notes/:id (UC-03)
func GetNoteByID(c *gin.Context) {
	var note models.Note
	noteID := c.Param("id")

	if err := database.DB.Preload("Tags").First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	attachRatingSummary(&note)

	var media []models.NoteMedia
	database.DB.Where("note_id = ?", noteID).Find(&media)

	var reviews []models.Review
	database.DB.Where("note_id = ?", noteID).Find(&reviews)

	type ReviewWithReplies struct {
		models.Review
		Replies []models.ReviewReply `json:"replies"`
	}
	var fullReviews []ReviewWithReplies

	for _, r := range reviews {
		var replies []models.ReviewReply
		database.DB.Where("review_id = ?", r.ReviewID).Find(&replies)
		fullReviews = append(fullReviews, ReviewWithReplies{Review: r, Replies: replies})
	}

	var dev models.User
	sellerName := "未知"
	if err := database.DB.First(&dev, note.SellerID).Error; err == nil {
		sellerName = dev.Username
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"note":           note,
			"seller_name": sellerName,
			"media":          media,
			"tags":           note.Tags,
			"reviews":        fullReviews,
		},
	})
}
