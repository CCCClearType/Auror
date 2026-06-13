package controllers

import (
	"net/http"
	"auror_vapor_backend/database"
	"auror_vapor_backend/models"

	"github.com/gin-gonic/gin"
)

// SuspendUser handles PUT /api/admin/users/:id/suspend
func SuspendUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.Permission == "DELETED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot modify suspension status of a deleted user"})
		return
	}

	// Toggle suspension
	if user.Permission == "ACTIVE" {
		user.Permission = "DEACTIVE"
	} else {
		user.Permission = "ACTIVE"
	}

	database.DB.Save(&user)
	c.JSON(http.StatusOK, gin.H{"message": "User account has been suspended"})
}

// DeleteUser handles DELETE /api/admin/users/:id
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Permission = "DELETED"
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	// Cascade soft-delete: if this user is a seller, take down all their notes
	if user.Role == "SELLER" {
		database.DB.Model(&models.Note{}).Where("seller_id = ?", user.UserID).Update("status", "TAKEN_DOWN")

		// Cascade revoke licenses for all notes owned by this seller
		var sellerNotes []uint
		database.DB.Model(&models.Note{}).Where("seller_id = ?", user.UserID).Pluck("note_id", &sellerNotes)
		if len(sellerNotes) > 0 {
			database.DB.Model(&models.NoteLicense{}).Where("note_id IN ?", sellerNotes).Update("status", "REVOKED")
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "User completely removed"})
}

// ChangeUserRole handles PUT /api/admin/users/:id/role
func ChangeUserRole(c *gin.Context) {
	userID := c.Param("id")

	var input struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Model(&models.User{}).Where("user_id = ?", userID).Update("role", input.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}

// AdminDeleteNote handles DELETE /api/admin/notes/:id
func AdminDeleteNote(c *gin.Context) {
	noteID := c.Param("id")
	if err := database.DB.Model(&models.Note{}).Where("note_id = ?", noteID).Update("status", "TAKEN_DOWN").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to take down note"})
		return
	}

	// Cascade revoke all licenses for this note
	database.DB.Model(&models.NoteLicense{}).Where("note_id = ?", noteID).Update("status", "REVOKED")

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully by Admin"})
}
