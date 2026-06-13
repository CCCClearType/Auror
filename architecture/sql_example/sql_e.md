# e. 進階查詢 II (使用 ORDER BY, IN, MAX/MIN/AVG/SUM/COUNT, GROUP BY, HAVING 等)

本文件收錄系統中使用到的第二階層進階 SQL 語法，包含群組聚合運算 (Aggregation)、排序、以及清單包含判斷 (`IN`)。

---

### 1. 筆記評分統計 (使用 COUNT, AVG 與 CASE WHEN)
- **說明**：當買家新增或刪除筆記評論時，系統會動態觸發此查詢，針對該款筆記的所有可見評論 (`VISIBLE`) 進行統計。利用 `COUNT` 算出總評論數，再利用 `AVG` 搭配 `CASE WHEN` 將正負評轉換為具體的 0.0 ~ 5.0 平均分數，最後將這個分數存回 `notes.overall_rating`。
- **對應功能**：新增/刪除評論時觸發的重算邏輯 (內部呼叫 `refreshStoredNoteRating`)
- **Go 實作 (GORM)**：
  ```go
  database.DB.Model(&models.Review{}).
      Where("note_id = ? AND status = 'VISIBLE'", noteID).
      Select("COUNT(*) as total_reviews, COALESCE(AVG(CASE WHEN attitude = 'POSITIVE' THEN 5.0 ELSE 1.0 END), 0) as average_rating").
      Row().Scan(&totalReviews, &averageRating)
  ```
- **原生 SQL 語法**：
  ```sql
  SELECT 
    COUNT(*) as total_reviews, 
    COALESCE(AVG(CASE WHEN attitude = 'POSITIVE' THEN 5.0 ELSE 1.0 END), 0) as average_rating 
  FROM reviews 
  WHERE note_id = 42 AND status = 'VISIBLE';
  ```

### 2. 商店首頁的分群與自訂排序 (使用 GROUP BY 與 ORDER BY)
- **說明**：在商店查詢時，為了避免因為 JOIN 多個科目導致同一款筆記重複出現，必須使用 `GROUP BY notes.note_id` 將結果合併。同時也利用 `ORDER BY` 實現依照價格由高至低或發售日期排列的功能。
- **對應 API**：`GET /api/notes?sort={sort_type}`
- **Go 實作 (GORM)**：
  ```go
  query = query.Group("notes.note_id")
  if sort == "price_desc" {
      query = query.Order("notes.price DESC")
  } else {
      query = query.Order("notes.release_date DESC")
  }
  ```
- **原生 SQL 語法**：
  ```sql
  SELECT notes.* FROM notes 
  LEFT JOIN note_tags ON note_tags.note_id = notes.note_id
  WHERE notes.status = 'ACTIVE' 
  GROUP BY notes.note_id
  ORDER BY notes.price DESC;
  ```

### 3. 結帳前的購物車商品清單校驗 (使用 IN)
- **說明**：當買家按下結帳按鈕，系統會將購物車內的多個 `note_id` 集合成一個陣列，並使用 `IN` 語法一次性向資料庫拉出所有筆記最新的價格與狀態，以防在結帳瞬間有筆記被下架或改價。
- **對應 API**：`POST /api/shopping/checkout`
- **Go 實作 (GORM)**：
  ```go
  var notes []models.Note
  database.DB.Where("note_id IN ?", noteIDs).Find(&notes)
  ```
- **原生 SQL 語法**：
  ```sql
  SELECT * FROM notes 
  WHERE note_id IN (14, 25, 42, 58);
  ```

### 4. 賣家總銷售額與銷量統計 (使用 SUM, COUNT)
- **說明**：賣家在查看某款筆記的統計數據時，系統會 JOIN 交易明細表，並利用聚合函數 `SUM` 計算總營業額、`COUNT` 統計總共賣出的套數。
- **對應 API**：`GET /api/seller/notes/:id/stats`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Table("transaction_items ti").
      Select("COUNT(ti.item_id) as total_sales_count, COALESCE(SUM(ti.purchase_price), 0) as total_revenue").
      Joins("JOIN notes g ON ti.note_id = g.note_id").
      Where("g.seller_id = ? AND g.note_id = ?", sellerID, noteID).
      Row().Scan(&stats.TotalSalesCount, &stats.TotalRevenue)
  ```
- **原生 SQL 語法**：
  ```sql
  SELECT 
    COUNT(ti.item_id) as total_sales_count,
    COALESCE(SUM(ti.purchase_price), 0) as total_revenue
  FROM transaction_items ti
  JOIN notes g ON ti.note_id = g.note_id
  WHERE g.seller_id = 5 AND g.note_id = 42;
  ```

### 5. 後台使用者列表排序 (ORDER BY)
- **說明**：管理員檢視全站使用者名單時，系統預設將最新註冊的帳號排在最前面，方便管理與監控近期湧入的會員。
- **對應 API**：`GET /api/admin/users`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Order("registration_date DESC").Find(&users)
  ```
- **原生 SQL 語法**：
  ```sql
  SELECT * FROM users ORDER BY registration_date DESC;
  ```

### 6. 退款紀錄列表排序 (ORDER BY 多欄位)
- **說明**：客服 (CSR) 介面需要列出退款申請，為確保客服第一時間看到亟需處理的項目，會進行多重欄位排序。此語法會先將 `status` 為 `PENDING` (待處理) 的單子排前面，並且依照申請時間正序排列 (最老的待處理單優先)。
- **對應 API**：`GET /api/csr/refunds`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Order("status DESC").Order("created_at ASC").Find(&refunds)
  ```
- **原生 SQL 語法**：
  ```sql
  SELECT * FROM refund_requests ORDER BY status DESC, created_at ASC;
  ```
