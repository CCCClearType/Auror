# c. 使用三個資料表的查詢 (Three-Table JOIN Queries)

本文件列出系統中橫跨三個或以上資料表的複雜關聯查詢 (JOIN)。這些查詢通常發生在多對多關聯、或是一對多對多的深層結構中。

---

### 1. 透過特定分類標籤搜尋筆記 (多對多關聯)
- **說明**：買家點擊某個學期 (例如 115-2) 來找筆記時，由於這是一個多對多關係，必須先從 `notes` 表 JOIN 中介表 `note_tags`，再 JOIN 科目主檔 `tags` 來做名稱與類型的比對。
- **對應 API**：`GET /api/notes?semester=115-2`
- **Go 實作 (GORM)**：
  ```go
  query = query.
      Joins("JOIN note_tags filter_sem_tags ON filter_sem_tags.note_id = notes.note_id").
      Joins("JOIN tags filter_sem ON filter_sem.tag_id = filter_sem_tags.tag_id").
      Where("filter_sem.tag_name ILIKE ? AND filter_sem.tag_type = 'SEMESTER'", semester)
  ```
- **原生 SQL 語法 (連續 INNER JOIN)**：
  ```sql
  SELECT notes.* 
  FROM notes
  JOIN note_tags ON note_tags.note_id = notes.note_id
  JOIN tags ON tags.tag_id = note_tags.tag_id
  WHERE notes.status = 'ACTIVE' 
    AND tags.tag_name ILIKE '115-2'
    AND tags.tag_type = 'SEMESTER';
  ```

### 2. 查詢買家的筆記庫並包含筆記封面圖
- **說明**：買家的筆記庫紀錄了買過的筆記 (`note_licenses`)。為了在前端渲染出筆記圖示，必須先串接 `notes` 主檔，再串接 `note_media` 取出附屬的圖片，總共橫跨三個表。
- **對應 API**：`GET /api/protected/library`
- **Go 實作 (GORM)**：
  ```go
  var licenses []models.NoteLicense
  database.DB.Preload("Note.Media").Where("user_id = ? AND status = 'ACTIVE'", userID).Find(&licenses)
  ```
- **原生 SQL 語法 (JOIN + LEFT JOIN 等效邏輯)**：
  ```sql
  SELECT note_licenses.license_id, notes.title, note_media.file_url 
  FROM note_licenses 
  JOIN notes ON note_licenses.note_id = notes.note_id 
  LEFT JOIN note_media ON notes.note_id = note_media.note_id 
  WHERE note_licenses.user_id = 5 
    AND note_licenses.status = 'ACTIVE';
  ```

### 3. 查詢歷史訂單明細與對應的筆記名稱
- **說明**：這是標準的電商一對多對一三表關聯。從主訂單表 (`transactions`) 關聯出底下的購物明細項目 (`transaction_items`)，再關聯到筆記檔案 (`notes`) 來顯示買家到底買了哪些筆記。
- **對應 API**：`GET /api/protected/transactions`
- **Go 實作 (GORM)**：
  ```go
  var transactions []models.Transaction
  database.DB.Preload("Items.Note").Where("user_id = ?", userID).Order("created_at DESC").Find(&transactions)
  ```
- **原生 SQL 語法 (連續 JOIN 等效邏輯)**：
  ```sql
  SELECT transactions.transaction_id, transactions.created_at, 
         transaction_items.purchase_price, notes.title
  FROM transactions
  JOIN transaction_items ON transactions.transaction_id = transaction_items.transaction_id
  JOIN notes ON transaction_items.note_id = notes.note_id
  WHERE transactions.user_id = 5
  ORDER BY transactions.created_at DESC;
  ```

### 4. 下載筆記檔案前之授權與實體檔案驗證
- **說明**：當買家點擊下載筆記時，系統必須橫跨三個表進行確認：1) 買家真的擁有這款筆記 (`note_licenses`)，2) 該筆記沒被硬刪除 (`notes`)，3) 該筆記有上傳對應的實體安裝檔 (`note_media`) 供下載。
- **對應 API**：`GET /api/protected/library/:note_id/download`
- **Go 實作 (GORM)**：
  ```go
  var license models.NoteLicense
  // 驗證授權 (Table 1: note_licenses, Table 2: notes)
  database.DB.Joins("Note").Where("user_id = ? AND note_licenses.note_id = ? AND note_licenses.status = 'ACTIVE'", userID, noteID).First(&license)

  var media models.NoteMedia
  // 尋找實體檔案 (Table 3: note_media)
  database.DB.Where("note_id = ? AND media_type = 'NOTE_FILE'", noteID).First(&media)
  ```
- **原生 SQL 語法 (邏輯上等同三表驗證 JOIN)**：
  ```sql
  SELECT note_licenses.license_id, notes.note_id, note_media.file_url 
  FROM note_licenses
  JOIN notes ON note_licenses.note_id = notes.note_id
  JOIN note_media ON notes.note_id = note_media.note_id
  WHERE note_licenses.user_id = 5 
    AND note_licenses.note_id = 42 
    AND note_licenses.status = 'ACTIVE'
    AND note_media.media_type = 'NOTE_FILE';
  ```
