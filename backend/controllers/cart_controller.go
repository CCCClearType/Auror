package controllers

import (
	"net/http"
	"auror_vapor_backend/database"
	"auror_vapor_backend/models"

	"github.com/gin-gonic/gin"
)

type AddCartInput struct {
	NoteID uint `json:"note_id" binding:"required"`
}

// GetCart handles GET /api/protected/cart
func GetCart(c *gin.Context) {
	// 1. Get user_id from the AuthMiddleware
	userID, _ := c.Get("user_id")

	// 2. Fetch all cart items for this user, and Preload the "Note" details
	var cartItems []models.ShoppingCart
	if err := database.DB.Preload("Note").Preload("Note.Media").Where("user_id = ?", userID).Find(&cartItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": cartItems})
}

// AddToCart handles POST /api/protected/cart
func AddToCart(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	userID := uint(userIDFloat.(float64)) // JWT numbers are parsed as float64

	var input AddCartInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Check if note already in cart
	var existing models.ShoppingCart
	if err := database.DB.Where("user_id = ? AND note_id = ?", userID, input.NoteID).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Note already in cart"})
		return
	}

	// 1.5 Check if the user already OWNS the note (has an ACTIVE license)
	var license models.NoteLicense
	if err := database.DB.Where("user_id = ? AND note_id = ? AND status = ?", userID, input.NoteID, "ACTIVE").First(&license).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You already own this note"})
		return
	}

	// 1.8 Check if the note is TAKEN_DOWN
	var note models.Note
	if err := database.DB.First(&note, input.NoteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	if note.Status != "ACTIVE" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This note is not available for purchase"})
		return
	}

	// 2. Add to cart
	cartItem := models.ShoppingCart{
		UserID: userID,
		NoteID: input.NoteID,
	}

	if err := database.DB.Create(&cartItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note added to cart successfully"})
}

// RemoveFromCart handles DELETE /api/protected/cart/:note_id
func RemoveFromCart(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	userID := uint(userIDFloat.(float64))
	noteID := c.Param("note_id")

	if err := database.DB.Where("user_id = ? AND note_id = ?", userID, noteID).Delete(&models.ShoppingCart{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note removed from cart"})
}
