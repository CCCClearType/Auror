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

func attachClassicStatus(note *models.Note) {
	isOldEnough := false
	for _, tag := range note.Tags {
		if tag.TagType == "SEMESTER" {
			if tag.TagName <= "111-2" {
				isOldEnough = true
			}
			break
		}
	}

	if !isOldEnough {
		note.IsClassic = false
		return
	}

	var salesCount int64
	database.DB.Table("transaction_items").Where("note_id = ?", note.NoteID).Count(&salesCount)
	note.IsClassic = salesCount >= 15
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
	semester := c.Query("semester")
	subject := c.Query("subject")
	teacher := c.Query("teacher")
	department := c.Query("department")
	courseType := c.Query("course_type")
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
			Joins("JOIN note_tags filter_general_tags ON filter_general_tags.note_id = notes.note_id").
			Joins("JOIN tags filter_general ON filter_general.tag_id = filter_general_tags.tag_id").
			Where("filter_general.tag_name ILIKE ?", tag)
	}
	if semester != "" {
		query = query.
			Joins("JOIN note_tags filter_sem_tags ON filter_sem_tags.note_id = notes.note_id").
			Joins("JOIN tags filter_sem ON filter_sem.tag_id = filter_sem_tags.tag_id").
			Where("filter_sem.tag_name ILIKE ? AND filter_sem.tag_type = 'SEMESTER'", semester)
	}
	if subject != "" {
		query = query.
			Joins("JOIN note_tags filter_sub_tags ON filter_sub_tags.note_id = notes.note_id").
			Joins("JOIN tags filter_sub ON filter_sub.tag_id = filter_sub_tags.tag_id").
			Where("filter_sub.tag_name ILIKE ? AND filter_sub.tag_type = 'SUBJECT'", subject)
	}
	if teacher != "" {
		query = query.
			Joins("JOIN note_tags filter_tea_tags ON filter_tea_tags.note_id = notes.note_id").
			Joins("JOIN tags filter_tea ON filter_tea.tag_id = filter_tea_tags.tag_id").
			Where("filter_tea.tag_name ILIKE ? AND filter_tea.tag_type = 'TEACHER'", teacher)
	}
	if department != "" {
		query = query.
			Joins("JOIN note_tags filter_dep_tags ON filter_dep_tags.note_id = notes.note_id").
			Joins("JOIN tags filter_dep ON filter_dep.tag_id = filter_dep_tags.tag_id").
			Where("filter_dep.tag_name ILIKE ? AND filter_dep.tag_type = 'DEPARTMENT'", department)
	}
	if courseType != "" {
		query = query.
			Joins("JOIN note_tags filter_ct_tags ON filter_ct_tags.note_id = notes.note_id").
			Joins("JOIN tags filter_ct ON filter_ct.tag_id = filter_ct_tags.tag_id").
			Where("filter_ct.tag_name ILIKE ? AND filter_ct.tag_type = 'COURSE_TYPE'", courseType)
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
		attachClassicStatus(&notes[i])
		var seller models.User
		if err := database.DB.Select("username").First(&seller, notes[i].SellerID).Error; err == nil {
			notes[i].SellerName = seller.Username
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
	attachClassicStatus(&note)

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

	var seller models.User
	sellerName := "未知"
	if err := database.DB.First(&seller, note.SellerID).Error; err == nil {
		sellerName = seller.Username
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
