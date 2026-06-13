# 10. 筆記庫與閱讀下載 (Note Library & Play)

本文件描述買家購買筆記後，在「個人筆記庫 (`library.html`)」中瀏覽已購筆記、下載檔案以及啟動筆記的完整業務邏輯。由後端 `library_controller.go` 負責處理。

---

## 1. 筆記庫列表 (Library Listing)

- **起點**：登入的買家點擊導覽列進入「筆記庫 (`library.html`)」。
- **流程 (`GET /api/protected/library`)**：
  1. 系統透過 JWT 確認使用者身分 (`userID`)。
  2. 後端查詢 `note_licenses` 表，過濾條件為：
     - `user_id = ?`
     - `status = 'ACTIVE'` (確保只有處於生效狀態的授權才會顯示)。
  3. 透過 GORM Preload 帶出關聯的 `Note` 主檔與 `Note.Media` (包含封面圖等)，一併回傳給前端。
- **防護機制**：
  - 若買家的筆記授權被管理員撤銷 (`REVOKED`) 或因退款被凍結 (`FROZEN`)，該筆記將**自動從筆記庫中消失**，前端無法取得該資料。
- **終點**：買家在頁面上看到網格排列的已擁有筆記清單。

---

## 2. 下載筆記檔案 (Download Note File)

- **起點**：買家在筆記庫中，針對特定筆記點擊「下載」按鈕。
- **流程 (`GET /api/protected/library/:note_id/download`)**：
  1. **授權驗證**：後端嚴格檢查 `note_licenses` 中，該買家對該 `note_id` 是否擁有 `status = 'ACTIVE'` 的紀錄。若無授權，直接回傳 HTTP 403 Forbidden。
  2. **尋找檔案實體**：在 `note_media` 表中尋找 `note_id = ?` 且 `media_type = 'note_file'` 的紀錄，取得 `file_url`。
  3. **路徑安全防護 (Path Traversal Prevention)**：
     - 系統要求 `file_url` 必須以 `/downloads/` 開頭。
     - 系統會移除前綴並透過 `filepath.Clean` 清理路徑。
     - **關鍵防護**：若清理後的路徑包含 `..` (回退目錄) 或以 `/` 開頭，後端會立即阻擋並回傳 `Invalid note file path`，防止駭客利用目錄遍歷漏洞讀取伺服器敏感檔案。
  4. **回傳實體檔案**：安全驗證通過後，將路徑映射至伺服器實體目錄 `assets/note-files/...` 並將二進位檔案以 Attachment 形式串流回傳給前端。
- **終點**：瀏覽器觸發檔案下載。

---

## 3. 啟動筆記 (Play Note)

- **起點**：買家在筆記庫中，針對特定筆記點擊「閱讀」按鈕。
- **流程 (`GET /api/protected/library/:note_id/play`)**：
  1. **授權驗證**：如同下載功能，後端嚴格檢查 `note_licenses` 中該買家對該筆記是否有 `ACTIVE` 狀態的授權。
  2. **核發憑證**：驗證通過後，後端會模擬回傳一組啟動憑證 (Token)，例如 `mock-play-token-12345`。
- **終點**：前端收到成功訊息與憑證後，以模擬方式彈出提示「啟動成功」，象徵筆記用戶端已接手該憑證並開始執行筆記。
