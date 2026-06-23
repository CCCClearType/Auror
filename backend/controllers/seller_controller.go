package controllers

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"auror_vapor_backend/database"
	"auror_vapor_backend/models"

	"github.com/gin-gonic/gin"
)

type UploadNoteInput struct {
	Title    string  `json:"title" binding:"required"`
	Price    float64 `json:"price" binding:"min=0"`
	Desc     string  `json:"desc"`
}

type UpdateNoteInput struct {
	Price float64 `json:"price" binding:"min=0"`
	Desc  string  `json:"desc"`
}

// GetSellerNotes handles GET /api/seller/notes
func GetSellerNotes(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	sellerID := uint(userIDFloat.(float64))

	var notes []models.Note
	query := database.DB.Preload("Media").Preload("Tags")
	if role, _ := c.Get("role"); role != "ADMIN" {
		query = query.Where("seller_id = ?", sellerID)
	}

	if err := query.Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch seller notes"})
		return
	}

	for i := range notes {
		for j := range notes[i].Media {
			physicalPath := getPhysicalPath(notes[i].Media[j].FileURL)
			if physicalPath != "" {
				if info, err := os.Stat(physicalPath); err == nil {
					notes[i].Media[j].FileSize = info.Size()
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": notes})
}

// UploadNote handles POST /api/seller/notes
func UploadNote(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	sellerID := uint(userIDFloat.(float64))

	var input UploadNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note := models.Note{
		SellerID:    sellerID,
		Title:       input.Title,
		Description: input.Desc,
		Price:       input.Price,
		Status:      "DRAFT",
	}

	// Insert new note into the database
	if err := database.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload note"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Note uploaded successfully",
		"note":    note,
	})
}

// PublishNote handles PUT /api/seller/notes/:id/publish
func PublishNote(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	sellerID := uint(userIDFloat.(float64))
	noteID := c.Param("id")

	var note models.Note
	// Preload Tags to check count
	if err := database.DB.Preload("Tags").First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if role, _ := c.Get("role"); role != "ADMIN" && note.SellerID != sellerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You can only publish your own notes"})
		return
	}

	if len(note.Tags) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Note must have at least 1 tag to be published"})
		return
	}

	hasSemester := false
	for _, t := range note.Tags {
		if t.TagType == "SEMESTER" {
			hasSemester = true
			break
		}
	}

	if !hasSemester {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Note must have at least 1 semester tag"})
		return
	}

	if err := database.DB.Model(&note).Update("status", "ACTIVE").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note published successfully"})
}

// UpdateNote handles PUT /api/seller/notes/:id
func UpdateNote(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	sellerID := uint(userIDFloat.(float64))
	noteID := c.Param("id")

	var input UpdateNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var note models.Note
	if err := database.DB.First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if role, _ := c.Get("role"); role != "ADMIN" && note.SellerID != sellerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You can only edit your own notes"})
		return
	}

	if err := database.DB.Model(&note).Updates(map[string]interface{}{
		"price":       input.Price,
		"description": input.Desc,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note"})
		return
	}

	database.DB.Preload("Media").First(&note, note.NoteID)
	c.JSON(http.StatusOK, gin.H{"message": "Note updated successfully", "note": note})
}

// DeleteNote handles DELETE /api/seller/notes/:id
func DeleteNote(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	sellerID := uint(userIDFloat.(float64))
	noteID := c.Param("id")

	// 1. Find the note first
	var note models.Note
	if err := database.DB.First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	// 2. IMPORTANT SECURITY CHECK: Ensure the seller trying to delete it is the owner
	if note.SellerID != sellerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You can only delete your own notes"})
		return
	}

	// 3. Delete the note (Soft delete by setting status to TAKEN_DOWN)
	if err := database.DB.Model(&note).Update("status", "TAKEN_DOWN").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to take down note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

// UploadMedia handles POST /api/seller/notes/:id/media
// Accepts multipart/form-data with fields:
//   - file       : the file to upload
//   - media_type : "media" (image, default) or "note_file"
//
// Storage paths:
//
//	media     → assets/images/{note_id}/{sha256}.{ext}      served at /media/images/{note_id}/{sha256}.{ext}
//	note_file → assets/note-files/{note_id}/{original_name} served at /downloads/{note_id}/{original_name}
func UploadMedia(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	sellerID := uint(userIDFloat.(float64))
	noteID := c.Param("id")

	var note models.Note
	if err := database.DB.First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if note.SellerID != sellerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You can only upload media for your own notes"})
		return
	}

	// Limit request body to slightly more than 40MB to accommodate multipart form overhead
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 41<<20)

	// Parse the uploaded file
	fileHeader, err := c.FormFile("file")
	if err != nil {
		if err.Error() == "http: request body too large" {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "檔案大小超過 40MB 限制"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing file field"})
		return
	}

	if fileHeader.Size > 40*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "檔案大小超過 40MB 限制"})
		return
	}

	totalSize, err := getNoteMediaTotalSize(noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate total media size"})
		return
	}

	if totalSize+fileHeader.Size > 50*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "單一筆記上傳素材大小總額不能超過 50MB"})
		return
	}

	mediaType := c.DefaultPostForm("media_type", "media")

	// Open and read file bytes
	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
		return
	}

	var dirPath, fileName, fileURL string

	if mediaType == "note_file" {
		// note_file: store under assets/note-files/{note_id}/{original_filename}
		// Sanitize the original filename (strip any directory components)
		originalName := filepath.Base(fileHeader.Filename)
		if originalName == "." || originalName == "" {
			originalName = "note_file.bin"
		}
		dirPath = filepath.Join("assets", "note-files", noteID)
		fileName = originalName
		fileURL = fmt.Sprintf("/downloads/%s/%s", noteID, fileName)
	} else {
		// media (image): store under assets/images/{note_id}/{sha256}.{ext}
		sum := sha256.Sum256(fileBytes)
		hashHex := fmt.Sprintf("%x", sum)
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if ext == "" {
			ext = ".bin"
		}
		dirPath = filepath.Join("assets", "images", noteID)
		fileName = hashHex + ext
		fileURL = fmt.Sprintf("/media/images/%s/%s", noteID, fileName)
	}

	// Ensure directory exists
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	// Write file to disk
	destPath := filepath.Join(dirPath, fileName)
	if err := os.WriteFile(destPath, fileBytes, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	media := models.NoteMedia{
		NoteID:    note.NoteID,
		FileURL:   fileURL,
		MediaType: mediaType,
	}

	if err := database.DB.Create(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save media record"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Media uploaded successfully",
		"data":     media,
		"file_url": fileURL,
	})
}

// DeleteMedia handles DELETE /api/seller/notes/:id/media/:media_id
// Also removes the physical file from disk.
func DeleteMedia(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	sellerID := uint(userIDFloat.(float64))
	noteID := c.Param("id")
	mediaID := c.Param("media_id")

	var note models.Note
	if err := database.DB.First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if note.SellerID != sellerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You can only manage your own notes"})
		return
	}

	// Fetch the media record first so we can remove the file
	var media models.NoteMedia
	if err := database.DB.Where("media_id = ? AND note_id = ?", mediaID, note.NoteID).First(&media).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	// Resolve the physical file path from the stored URL
	physicalPath := getPhysicalPath(media.FileURL)

	// Delete DB record
	if err := database.DB.Delete(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete media"})
		return
	}

	// Best-effort physical file removal (ignore error — file may already be gone)
	if physicalPath != "" {
		_ = os.Remove(physicalPath)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Media deleted successfully"})
}

// GetNoteStats handles GET /api/protected/seller/notes/:id/stats
func GetNoteStats(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	sellerID := uint(userIDFloat.(float64))
	noteID := c.Param("id")

	var note models.Note
	if err := database.DB.First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if note.SellerID != sellerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You can only view stats for your own notes"})
		return
	}

	// Advanced Aggregation Query (SUM and COUNT)
	var result struct {
		TotalSales   int     `json:"total_sales"`
		TotalRevenue float64 `json:"total_revenue"`
	}

	database.DB.Table("transaction_items").
		Select("count(*) as total_sales, COALESCE(sum(purchase_price), 0) as total_revenue").
		Where("note_id = ?", note.NoteID).
		Scan(&result)

	c.JSON(http.StatusOK, gin.H{"stats": result})
}

// GetTags handles GET /api/tags (Public)
func GetTags(c *gin.Context) {
	type TagWithCount struct {
		TagID     uint   `json:"tag_id"`
		TagName   string `json:"tag_name"`
		TagType   string `json:"tag_type"`
		NoteCount int    `json:"note_count"`
	}
	var tags []TagWithCount
	if err := database.DB.Raw(`
		SELECT t.tag_id, t.tag_name, t.tag_type, COUNT(nt.note_id) AS note_count
		FROM tags t
		LEFT JOIN note_tags nt ON t.tag_id = nt.tag_id
		GROUP BY t.tag_id, t.tag_name, t.tag_type
		ORDER BY t.tag_name DESC
	`).Scan(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tags})
}

// CreateTag handles POST /api/tags (Seller)
func CreateTag(c *gin.Context) {
	var input struct {
		TagName string `json:"tag_name" binding:"required"`
		TagType string `json:"tag_type"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tagName := strings.TrimSpace(input.TagName)
	if tagName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag name is required"})
		return
	}

	var existing models.Tag
	if err := database.DB.Where("LOWER(tag_name) = LOWER(?)", tagName).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Tag already exists", "data": existing})
		return
	}

	tagType := "GENERAL"
	if input.TagType != "" {
		tagType = strings.ToUpper(input.TagType)
	}

	if tagType == "SEMESTER" {
		c.JSON(http.StatusForbidden, gin.H{"error": "學期標籤由系統統一管理，無法自行新增！"})
		return
	}

	tag := models.Tag{TagName: tagName, TagType: tagType}
	if err := database.DB.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tag (might already exist)"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Tag created successfully", "data": tag})
}

// AddTagToNote handles POST /api/protected/seller/notes/:id/tags
func AddTagToNote(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	sellerID := uint(userIDFloat.(float64))
	noteID := c.Param("id")

	var note models.Note
	if err := database.DB.First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if role, _ := c.Get("role"); role != "ADMIN" && note.SellerID != sellerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Not your note"})
		return
	}

	var input struct {
		TagID uint `json:"tag_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	noteTag := models.NoteTag{NoteID: note.NoteID, TagID: input.TagID}
	if err := database.DB.Where("note_id = ? AND tag_id = ?", note.NoteID, input.TagID).FirstOrCreate(&noteTag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add tag to note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag added to note"})
}

// RemoveTagFromNote handles DELETE /api/protected/seller/notes/:id/tags/:tag_id
func RemoveTagFromNote(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	sellerID := uint(userIDFloat.(float64))
	noteID := c.Param("id")
	tagID := c.Param("tag_id")

	var note models.Note
	if err := database.DB.First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if role, _ := c.Get("role"); role != "ADMIN" && note.SellerID != sellerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	if err := database.DB.Where("note_id = ? AND tag_id = ?", note.NoteID, tagID).Delete(&models.NoteTag{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove tag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag removed from note"})
}

func getNoteMediaTotalSize(noteID string) (int64, error) {
	var totalSize int64
	dirs := []string{
		filepath.Join("assets", "images", noteID),
		filepath.Join("assets", "note-files", noteID),
	}

	for _, dir := range dirs {
		err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return nil // Ignore if directory doesn't exist
				}
				return err
			}
			if !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
		if err != nil {
			return 0, err
		}
	}
	return totalSize, nil
}

func getPhysicalPath(fileURL string) string {
	if strings.HasPrefix(fileURL, "/media/images/") {
		return filepath.Join("assets", "images", strings.TrimPrefix(fileURL, "/media/images/"))
	}
	if strings.HasPrefix(fileURL, "/downloads/") {
		return filepath.Join("assets", "note-files", strings.TrimPrefix(fileURL, "/downloads/"))
	}
	return ""
}
