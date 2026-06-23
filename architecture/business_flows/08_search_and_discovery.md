# 8. 筆記搜尋與探索 (Search & Discovery)

本文件描述買家在商店首頁 (`index.html`) 如何透過前端的三大控制面板（頂部搜尋列、左側過濾面板、右側進階控制）進行過濾與搜尋，並詳述其對應的底層邏輯。

---

## 1. 頂部搜尋列 (Top Search Bar)

- **UI 位置**：頁面正上方、最醒目的橫向搜尋框。
- **功能**：關鍵字模糊搜尋 (Keyword Fuzzy Search)。買家輸入文字後點擊「搜尋」按鈕。點擊進入個別筆記時，前端會呼叫 `GET /api/notes/{id}` 顯示詳情。
- **底層邏輯**：
  1. 前端將關鍵字放入 `q` 參數，呼叫 `GET /api/notes?q={keyword}`。
  2. 後端組建 SQL 查詢，針對以下四個維度進行 `ILIKE` 模糊比對：
     - `notes.title` (筆記標題)
     - `notes.description` (筆記簡介)
     - `q_tags.tag_name` (筆記綁定的科目名稱)
     - `q_sellers.username` (賣家的名稱)
  3. 只要上述任一欄位包含關鍵字，該筆記就會被篩選出來。

---

## 2. 左側過濾面板 (Left Filter Panel)

- **UI 位置**：頁面左側的垂直區塊，包含「Tag 瀏覽」與「價格篩選」兩大區塊。
- **功能與邏輯**：
  - **Tag 瀏覽 (分類標籤過濾)**：
    - 前端會列出學期 (`SEMESTER`)、科目 (`SUBJECT`)、老師 (`TEACHER`)、開課系所 (`DEPARTMENT`) 等分類標籤（從 `/api/tags` 動態獲取）。
    - 側邊欄的標籤會透過 `note_count` 顯示該標籤目前關聯的筆記數量（例如：`115-2 (11)`），這個統計數字是由 `GET /api/tags` 透過 `LEFT JOIN note_tags` 計算得出的。
    - 當買家點擊特定標籤，前端會發送對應類型的查詢參數，例如：`?semester=115-2` 或 `?subject=微積分`。
    - 後端收到參數後，會透過 `JOIN note_tags` 與 `tags` 資料表，精準過濾出綁定該類別標籤的筆記。選擇「所有筆記」則會清除分類參數。
  - **賣家篩選 (Seller Filter)**：
    - 當買家從特定賣家頁面進入，或點擊賣家名稱時，發送 `?seller={username}`。
    - 後端會透過 `ILIKE` 執行「精確的字串比對」(不加上 `%` 萬用字元)，確保輸入 `dev_1` 不會錯誤搜出 `dev_10`。
  - **價格篩選 (Price Filter)**：
    - 提供預設的價格區間選項，例如：
      - 「免費下載」：發送 `?max_price=0`。
      - 「NT$ 300 以下」：發送 `?max_price=300`。
      - 「NT$ 301 - 900」：發送 `?min_price=301&max_price=900`。
    - 後端收到 `min_price` 與 `max_price` 後，會直接在 SQL 中加上 `notes.price >= ?` 與 `notes.price <= ?` 的條件。

---

## 3. 右側進階控制面板 (Right Advanced Controls)

- **UI 位置**：筆記列表右上角，包含一個「隱藏已購買」的開關 (Toggle) 與一個「排序方式」的下拉選單。
- **功能與邏輯**：
  - **隱藏已購買 (Hide Owned)**：
    - 這是一個 ON/OFF 切換開關。
    - 當開啟時，前端發送 `?hide_owned=true`，且必須在 Request Header 帶上使用者的 JWT Token。
    - 後端偵測到該參數與 Token 後，會自動利用子查詢 (`NOT EXISTS`) 到 `note_licenses` 檢查，**並額外比對 `notes.seller_id != 當前使用者`**。只要該買家擁有的授權是 `ACTIVE`，或是該筆記正是由登入者自己開發的，該筆記就會從搜尋結果中被剔除，讓買家專注於探索尚未擁有的新筆記。
  - **排序方式 (Sorting Dropdown)**：
    - 提供如「價格由低到高」或「價格由高到低」的選項。
    - 選擇後前端發送 `?sort=price_asc` 或 `?sort=price_desc`。
    - 後端將其轉換為 SQL 的 `ORDER BY notes.price ASC/DESC`。

---

## 4. 綜合搜尋與最底層防護

- **綜合查詢**：上述三個面板的操作可以**疊加使用**。例如，買家可以同時搜尋 "Action" (頂部)，勾選 "NT$ 300 以下" (左側)，並開啟 "隱藏已購買" (右側)。前端會將這些參數全部串接起來 (`?q=Action&max_price=300&hide_owned=true&sort=price_asc`) 一次發送給後端。
- **最底層防護**：無論買家如何組合過濾條件，後端的查詢產生器最底層一定會加上 `WHERE notes.status = 'ACTIVE'`。這確保了任何被管理員「強制下架」(`TAKEN_DOWN`) 或賣家自主下架的筆記，絕對不會出現在任何搜尋結果中。
