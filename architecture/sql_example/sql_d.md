# d. 進階查詢 I (使用 EXISTS, NOT EXISTS, NULL, UNION, >=, LIKE 等)

本文件收錄系統中使用到的第一階層進階 SQL 語法，包含模糊搜尋、數值範圍區間比對、以及利用子查詢 (Subquery) 進行存在性 (`EXISTS` / `NOT EXISTS`) 判斷的實際應用範例。

---

### 1. 使用 NOT EXISTS 過濾已購買的筆記
- **說明**：買家在瀏覽商店首頁時，可以勾選「隱藏已擁有的筆記」。此時後端會利用 `NOT EXISTS` 子查詢去檢查 `note_licenses` 中是否已經有這款筆記的購買紀錄，將買過的濾除。
- **對應 API**：`GET /api/notes?hide_owned=true`
- **Go 實作 (GORM)**：
  ```go
  query = query.Where("NOT EXISTS (SELECT 1 FROM note_licenses WHERE note_licenses.note_id = notes.note_id AND note_licenses.user_id = ? AND note_licenses.status = 'ACTIVE')", userID)
  ```
- **原生 SQL 語法**：
  ```sql
  SELECT * FROM notes 
  WHERE notes.status = 'ACTIVE' 
    AND NOT EXISTS (
      SELECT 1 FROM note_licenses 
      WHERE note_licenses.note_id = notes.note_id 
        AND note_licenses.user_id = 5 
        AND note_licenses.status = 'ACTIVE'
    );
  ```

### 2. 使用 ILIKE 進行多欄位模糊搜尋 (Keyword Search)
- **說明**：當買家在搜尋框輸入關鍵字時，系統必須同時去比對筆記標題、筆記描述、科目名稱以及賣家名稱。透過 `ILIKE` 能達到不分大小寫的模糊比對效果。
- **對應 API**：`GET /api/notes?q={keyword}`
- **Go 實作 (GORM)**：
  ```go
  keyword := "%" + q + "%"
  query = query.Where("notes.title ILIKE ? OR notes.description ILIKE ? OR filter_tags.tag_name ILIKE ? OR filter_sellers.username ILIKE ?", keyword, keyword, keyword, keyword)
  ```
- **原生 SQL 語法**：
  ```sql
  SELECT notes.* FROM notes
  LEFT JOIN note_tags ON note_tags.note_id = notes.note_id
  LEFT JOIN tags ON tags.tag_id = note_tags.tag_id
  LEFT JOIN users ON users.user_id = notes.seller_id
  WHERE notes.status = 'ACTIVE' AND (
    notes.title ILIKE '%戰神%' OR 
    notes.description ILIKE '%戰神%' OR 
    tags.tag_name ILIKE '%戰神%' OR 
    users.username ILIKE '%戰神%'
  );
  ```

### 3. 使用 >= 與 <= 進行價格區間過濾
- **說明**：商店支援透過設定最低與最高價格區間來過濾筆記清單，幫助買家找到符合預算的筆記。
- **對應 API**：`GET /api/notes?min_price=100&max_price=500`
- **Go 實作 (GORM)**：
  ```go
  query = query.Where("notes.price >= ?", minPrice)
  query = query.Where("notes.price <= ?", maxPrice)
  ```
- **原生 SQL 語法**：
  ```sql
  SELECT * FROM notes 
  WHERE status = 'ACTIVE' 
    AND price >= 100.00 
    AND price <= 500.00;
  ```

### 4. 後台管理員使用 ILIKE 搜尋使用者
- **說明**：管理員在後台想要尋找特定買家時，可以使用模糊搜尋同時比對 `username` 與 `email` 欄位。
- **對應 API**：`GET /api/admin/users?q={keyword}`
- **Go 實作 (GORM)**：
  ```go
  keyword := "%" + q + "%"
  database.DB.Where("username ILIKE ? OR email ILIKE ?", keyword, keyword).Order("registration_date DESC").Find(&users)
  ```
- **原生 SQL 語法**：
  ```sql
  SELECT * FROM users 
  WHERE username ILIKE '%test%' OR email ILIKE '%test%'
  ORDER BY registration_date DESC;
  ```

### 5. 確保審核通過的筆記才顯示 (使用 != 不等於)
- **說明**：這是隱含在所有前台顯示邏輯中的條件操作，確保被下架 (`TAKEN_DOWN`) 或仍是草稿 (`DRAFT`) 的筆記不會被一般會員搜尋到。
- **對應 API**：(所有前台 `GET /api/notes` 相關路由)
- **Go 實作 (GORM)**：
  ```go
  query = query.Where("notes.status != 'TAKEN_DOWN'")
  ```
- **原生 SQL 語法**：
  ```sql
  SELECT * FROM notes WHERE status != 'TAKEN_DOWN';
  ```
