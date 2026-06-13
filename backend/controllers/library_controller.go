package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"auror_vapor_backend/database"
	"auror_vapor_backend/models"

	"github.com/gin-gonic/gin"
)

// GetLibrary handles GET /api/protected/library
func GetLibrary(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	userID := uint(userIDFloat.(float64))

	var licenses []models.NoteLicense
	// Preload Note and Note.Media details, and only fetch ACTIVE licenses
	if err := database.DB.Preload("Note").Preload("Note.Media").Where("user_id = ? AND status = ?", userID, "ACTIVE").Find(&licenses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch library"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": licenses})
}

// GetWishlist handles GET /api/protected/wishlist
func GetWishlist(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	userID := uint(userIDFloat.(float64))

	var wishlist []models.WishList
	if err := database.DB.Preload("Note").Preload("Note.Media").Where("user_id = ?", userID).Find(&wishlist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wishlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wishlist})
}

// AddToWishlist handles POST /api/protected/wishlist
func AddToWishlist(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	userID := uint(userIDFloat.(float64))

	var input struct {
		NoteID uint `json:"note_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.WishList
	if err := database.DB.Where("user_id = ? AND note_id = ?", userID, input.NoteID).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Already in wishlist", "already_exists": true})
		return
	}

	var note models.Note
	if err := database.DB.First(&note, input.NoteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	// Removed TAKEN_DOWN check per user requirement: Wishlist is a simple mark state.

	wishItem := models.WishList{
		UserID: userID,
		NoteID: input.NoteID,
	}

	if err := database.DB.Create(&wishItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to wishlist (might already exist)"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Added to wishlist"})
}

// RemoveFromWishlist handles DELETE /api/protected/wishlist/:note_id
func RemoveFromWishlist(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	userID := uint(userIDFloat.(float64))
	noteID := c.Param("note_id")

	if err := database.DB.Where("user_id = ? AND note_id = ?", userID, noteID).Delete(&models.WishList{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from wishlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Removed from wishlist"})
}

// PlayNote handles GET /api/protected/library/:note_id/play
func PlayNote(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	userID := uint(userIDFloat.(float64))
	noteID := c.Param("note_id")

	var license models.NoteLicense
	if err := database.DB.Where("user_id = ? AND note_id = ? AND status = ?", userID, noteID, "ACTIVE").First(&license).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not own this note or the license is inactive"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note launched successfully", "auth_token": "mock-play-token-12345"})
}

// DownloadNote handles GET /api/protected/library/:note_id/download
func DownloadNote(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	userID := uint(userIDFloat.(float64))
	noteID := c.Param("note_id")

	var license models.NoteLicense
	if err := database.DB.Where("user_id = ? AND note_id = ? AND status = ?", userID, noteID, "ACTIVE").First(&license).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not own this note or the license is inactive"})
		return
	}

	var noteFile models.NoteMedia
	if err := database.DB.Where("note_id = ? AND media_type = ?", noteID, "note_file").First(&noteFile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No downloadable note file is available"})
		return
	}

	// Support both new path format (/downloads/{note_id}/{file}) and legacy (/downloads/{file})
	var fullPath string
	fileURL := noteFile.FileURL
	if strings.HasPrefix(fileURL, "/downloads/") {
		rel := strings.TrimPrefix(fileURL, "/downloads/")
		rel = filepath.Clean(rel)
		if strings.Contains(rel, "..") || strings.HasPrefix(rel, "/") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note file path"})
			return
		}
		fullPath = filepath.Join("assets", "note-files", rel)

		// Fallback for legacy DB entries: If file isn't found at the root, check inside the note_id subfolder
		if _, err := os.Stat(fullPath); os.IsNotExist(err) && !strings.Contains(rel, "/") && !strings.Contains(rel, "\\") {
			fallbackPath := filepath.Join("assets", "note-files", noteID, rel)
			if _, err2 := os.Stat(fallbackPath); err2 == nil {
				fullPath = fallbackPath
			}
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note file path"})
		return
	}

	c.FileAttachment(fullPath, filepath.Base(fullPath))
}
