package main

import (
	"log"
	"auror_vapor_backend/database"
	"auror_vapor_backend/models"
	"auror_vapor_backend/routes"
	"auror_vapor_backend/utils"
)

func main() {
	// Initialize database connection
	database.ConnectDB()

	// 臨時修復：將資料庫內明碼的假密碼轉換為真實的 Bcrypt 雜湊密碼
	// 讓預設測試帳號可以使用密碼 'admin' 登入
	var users []models.User
	database.DB.Find(&users)
	for _, u := range users {
		// 如果密碼不是以 $2a$ 開頭 (Bcrypt 格式)，就強制更新為 'admin' 的雜湊
		if len(u.PasswordHash) < 4 || u.PasswordHash[:4] != "$2a$" {
			hashed, _ := utils.HashPassword("admin")
			database.DB.Model(&u).Update("password_hash", hashed)
		}
	}

	// Move legacy frontend-hosted media paths to backend-hosted media/download URLs.
	database.DB.Exec("ALTER TABLE notes ADD COLUMN IF NOT EXISTS description TEXT DEFAULT ''")
	database.DB.Exec("UPDATE note_media SET file_url = ? WHERE note_id = 1 AND media_type = 'note_file'", "/downloads/1/data-structures-notes.txt")
	database.DB.Exec("UPDATE note_media SET file_url = ? WHERE note_id = 4 AND media_type = 'note_file'", "/downloads/4/calculus-guide.txt")
	database.DB.Exec("UPDATE note_media SET file_url = regexp_replace(file_url, '^https?://[^/]+', '') WHERE file_url ~ '^https?://[^/]+/media/images/' AND media_type <> 'note_file'")
	database.DB.Exec("UPDATE note_media SET file_url = REPLACE(file_url, '/assets/images/', '/media/images/') WHERE file_url LIKE '/assets/images/%' AND media_type <> 'note_file'")
	database.DB.Exec("UPDATE note_media SET file_url = regexp_replace(file_url, '^https?://[^/]+', '') WHERE file_url ~ '^https?://[^/]+/downloads/' AND media_type = 'note_file'")
	database.DB.Exec("UPDATE note_media SET file_url = ? WHERE file_url LIKE ?", "/media/images/protoss_12+8.png", "%protoss_knife.png")
	database.DB.Exec("UPDATE note_media SET file_url = ? WHERE file_url LIKE ?", "/media/images/protoss_cross.png", "%protoss_best_16.png")

	// Fix existing TAKEN_DOWN notes to have REVOKED licenses
	database.DB.Exec("UPDATE note_licenses SET status = 'REVOKED' FROM notes WHERE note_licenses.note_id = notes.note_id AND notes.status = 'TAKEN_DOWN'")

	// Setup Gin router
	r := routes.SetupRouter()

	// Start server on port 8000
	log.Println("Starting server on port 8000...")
	if err := r.Run(":8000"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
