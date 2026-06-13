# 2. 賣家筆記上架與管理 (Seller Note Management)

這份文件描述了具有 `SELLER` 權限的帳號，如何在 AurorNote 平台上發布、管理與下架自己的筆記。

---

## 1. 筆記草稿建立 (Create Note Draft)

- **起點**：賣家在 `seller_dashboard.html` 查看自己的筆記列表 (呼叫 `GET /api/seller/notes`)，並點擊「新增筆記」。
- **流程**：
  1. 輸入基本的 `Title` (標題)、`Price` (定價)、`Description` (支援 Markdown 格式的介紹)。
  2. 呼叫 `POST /api/seller/notes`。
  3. 後端驗證使用者身分是否為 `SELLER`，並將賣家 ID 綁定到該款筆記。
  4. 筆記建立完成，初始狀態預設為 `'DRAFT'` (草稿)，並在 `notes` 資料表生成唯一 `note_id`。此狀態下筆記不會出現在商店首頁。
- **終點**：畫面上顯示「建立成功」，自動跳轉進入 `edit_note.html?id={note_id}` 的進階編輯介面，賣家可繼續上傳素材與科目。

---

## 2. 筆記素材與檔案上傳 (Upload Media & Note Files)

- **起點**：在 `edit_note.html` 內的「上傳多媒體素材」與「上傳筆記檔案」區塊。
- **流程**：
  1. 賣家選擇圖片 (封面圖、宣傳圖)、影片 (預告片) 或是 ZIP/EXE 筆記主程式檔。
  2. 呼叫 `POST /api/seller/notes/:id/media` (使用 `multipart/form-data`)。
  3. 後端根據檔案類型 (`media_type`) 進行不同的實體儲存處理：
     - **圖片或影片 (`media` / `thumbnail`)**：會將檔案內容進行 SHA-256 雜湊 (Hash) 產生新的檔名，並存入 `assets/images/{note_id}/{hash}.{ext}`，避免檔名衝突。
     - **筆記檔案 (`note_file`)**：**不會進行 Hash**，而是保留賣家上傳的**原始檔名**，存入專屬資料夾 `assets/note-files/{note_id}/{original_filename}`。
  4. 實體路徑會轉換成對應的虛擬路由：圖片對應至 `/media/images/{note_id}/{hash}.{ext}`，筆記檔案則對應至受保護的下載路由 `/downloads/{note_id}/{original_filename}`，並寫入 `note_media` 表。
  5. **刪除素材**：若賣家欲刪除，可呼叫 `DELETE /api/seller/notes/:id/media/:media_id` 移除不要的素材。
- **終點**：前端收到新素材的 URL，並立刻渲染預覽圖或顯示檔案名稱。

---

## 3. 科目標理 (Tags Management)

- **起點**：賣家在編輯頁面想要為筆記加上科目 (例如：RPG, Action)。
- **流程**：
  1. 系統可以列出目前資料庫中所有的全域科目 (`GET /api/tags`)。
  2. 若無適合科目，賣家可呼叫 `POST /api/seller/tags` 創建新科目。
  3. 呼叫 `POST /api/seller/notes/:id/tags` 將 `tag_id` 與 `note_id` 綁定。這會在 `note_tags` (多對多關聯表) 中新增一筆紀錄。
  4. 若需解除科目，呼叫 `DELETE /api/seller/notes/:id/tags/:tag_id` 將其移除。
- **終點**：筆記分類變得精準，買家在首頁可以透過科目篩選出這款筆記。

---

## 4. 正式發佈上架 (Publish Note)

- **起點**：賣家在 `seller_dashboard.html` 看到草稿狀態的筆記，點擊「正式發佈」。
- **流程**：
  1. 呼叫 `PUT /api/seller/notes/:id/publish`。
  2. **【重要約束點】**：後端嚴格驗證該筆記**是否至少擁有 1 個科目 (Tag)**。
  3. 若條件不符，退回 HTTP 400，並提示賣家前往編輯頁面補齊資料。
  4. 驗證通過後，將筆記狀態由 `'DRAFT'` 更新為 `'ACTIVE'`。
- **終點**：筆記正式上架，出現在商店首頁，買家可開始瀏覽與購買。

---

## 5. 筆記資訊更新與自主下架 (Update & Take Down)

- **起點**：賣家想要修改價格、介紹，或是決定停止販售這款筆記。
- **流程 (更新資訊)**：
  - 呼叫 `PUT /api/seller/notes/:id`，傳入新的 `Price` 與 `Description`。
- **流程 (自主下架 Delete)**：
  1. 呼叫 `DELETE /api/seller/notes/:id`。
  2. **重要安全檢查**：後端會驗證這個 `note_id` 的 `SellerID` 是否與當前登入者相符，防止跨權限刪除。
  3. 通過驗證後，將該筆記的狀態 `notes.status` 設為 `'TAKEN_DOWN'` (軟刪除)。
  4. **與管理員下架的差異**：賣家自主下架時，**不會**連帶撤銷買家的 `note_licenses`。這是保障消費者的基本權益，也就是「買過的買家依然可以在筆記庫中找到它並下載閱讀，只是商店不再開放新買家購買」。
- **終點**：筆記在商店首頁隱藏，進入詳細頁面會顯示「此筆記已下架」，且無法加入購物車。

---

## 6. 筆記銷售數據統計 (Sales Stats)

- **起點**：賣家想要知道自己的筆記賣得如何。
- **流程**：
  - 呼叫 `GET /api/seller/notes/:id/stats`。
  - 後端會從 `transaction_items` 加總銷售數量與總收入。
- **終點**：在儀表板呈現銷售業績數據。
