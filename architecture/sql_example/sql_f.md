# f. 新增或刪除資料的操作 (INSERT, UPDATE, DELETE)

本文件列出系統進行資料異動 (Data Manipulation) 時使用的 SQL 操作。由於這涵蓋了系統中所有的狀態變更 API，以下依業務模組進行分類。

---

## 模組一：使用者與身分驗證 (User & Auth)

### 1. 買家註冊 (INSERT)
- **說明**：使用者註冊帳號時，系統會將密碼雜湊化後連同使用者名稱與信箱寫入資料庫，並取得自動遞增生成的 `user_id`。
- **對應 API**：`POST /api/auth/register`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Create(&user)
  ```
- **原生 SQL 語法**：
  ```sql
  INSERT INTO users (username, email, password_hash, role) 
  VALUES ('JohnDoe', 'john@test.com', '$2a$10$...', 'USERS') RETURNING user_id;
  ```

### 2. 更新個人檔案 (UPDATE)
- **說明**：使用者修改帳號資訊 (如使用者名稱、信箱) 時，系統會針對這些特定欄位進行更新。
- **對應 API**：`PUT /api/users/profile`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Save(&user)
  ```
- **原生 SQL 語法**：
  ```sql
  UPDATE users SET username = 'CoolPlayer', email = 'cool@test.com' WHERE user_id = 5;
  ```

---

## 模組二：商城與交易 (Store, Cart, Wishlist, Transactions)

### 3. 將筆記加入購物車 (INSERT)
- **說明**：當買家點擊加入購物車時，系統會在 `shopping_carts` 表寫入一筆對應的使用者與筆記關聯紀錄。
- **對應 API**：`POST /api/protected/cart`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Create(&cartItem)
  ```
- **原生 SQL 語法**：
  ```sql
  INSERT INTO shopping_carts (user_id, note_id) VALUES (5, 42);
  ```

### 4. 移除單一購物車商品 (DELETE)
- **說明**：買家將筆記移出購物車時，系統根據 `user_id` 與 `note_id` 的組合將該紀錄實體刪除。
- **對應 API**：`DELETE /api/protected/cart/:id`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Where("user_id = ? AND note_id = ?", userID, noteID).Delete(&models.ShoppingCart{})
  ```
- **原生 SQL 語法**：
  ```sql
  DELETE FROM shopping_carts WHERE user_id = 5 AND note_id = 42;
  ```

### 5. 清空購物車 (DELETE)
- **說明**：在結帳完成或買家手動清空時，系統一次性刪除該名買家在 `shopping_carts` 裡面的所有紀錄。
- **對應 API**：`DELETE /api/protected/cart`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Where("user_id = ?", userID).Delete(&models.ShoppingCart{})
  ```
- **原生 SQL 語法**：
  ```sql
  DELETE FROM shopping_carts WHERE user_id = 5;
  ```

### 6. 購物車結帳 (巨型 Transaction: INSERT + DELETE)
- **說明**：結帳是一個跨資料表的巨型交易，必須保證原子性。系統會建立主訂單、逐一建立子明細、配發筆記授權，最後再將使用者的購物車清空。
- **對應 API**：`POST /api/shopping/checkout`
- **Go 實作 (GORM)**：
  ```go
  tx := database.DB.Begin()
  tx.Create(&transaction)
  tx.Create(&transactionItem)
  tx.Create(&license)
  tx.Where("user_id = ?", userID).Delete(&models.ShoppingCart{})
  tx.Commit()
  ```
- **原生 SQL 語法**：
  ```sql
  WITH new_tx AS (
      INSERT INTO transactions (user_id, total_amount, receipt_number) 
      VALUES (4, 1200.00, 'REC-DEMO-0002') 
      RETURNING transaction_id
  ), new_item AS (
      INSERT INTO transaction_items (transaction_id, note_id, purchase_price) 
      SELECT transaction_id, 1, 1200.00 FROM new_tx 
      RETURNING item_id
  ), new_license AS (
      INSERT INTO note_licenses (user_id, note_id, transaction_item_id, status) 
      SELECT 4, 1, item_id, 'ACTIVE' FROM new_item
  )
  DELETE FROM shopping_carts WHERE user_id = 4;
  ```

### 7. 將筆記加入願望清單 (INSERT)
- **說明**：買家關注某款筆記時，系統純粹做一個標記，將其寫入 `wish_lists` 資料表中。
- **對應 API**：`POST /api/protected/wishlist`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Create(&wishlistItem)
  ```
- **原生 SQL 語法**：
  ```sql
  INSERT INTO wish_lists (user_id, note_id) VALUES (5, 42);
  ```

### 8. 移除願望清單 (DELETE)
- **說明**：買家取消關注或購買完成後，從 `wish_lists` 移除該關聯紀錄。
- **對應 API**：`DELETE /api/protected/wishlist/:id`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Where("user_id = ? AND note_id = ?", userID, noteID).Delete(&models.WishList{})
  ```
- **原生 SQL 語法**：
  ```sql
  DELETE FROM wish_lists WHERE user_id = 5 AND note_id = 42;
  ```

### 9. 買家申請退款 (INSERT)
- **說明**：買家提出退款要求時，系統會產生一張新的待處理退款單 (`PENDING`)，交由客服人員後續審核。
- **對應 API**：`POST /api/social/refunds`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Create(&refundRequest)
  ```
- **原生 SQL 語法**：
  ```sql
  INSERT INTO refund_requests (transaction_item_id, buyer_id, reason, status) 
  VALUES (105, 5, '不好玩', 'PENDING');
  ```

---

## 模組三：社群互動 (Social, Friends, Reviews)

### 10. 送出好友邀請 (INSERT)
- **說明**：買家想認識其他人時發出邀請，系統建立一筆狀態為 `PENDING` 的待確認好友關係。
- **對應 API**：`POST /api/social/friends/requests`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Create(&friendReq)
  ```
- **原生 SQL 語法**：
  ```sql
  INSERT INTO friendships (sender_id, receiver_id, status) VALUES (5, 10, 'PENDING');
  ```

### 11. 接受好友邀請 (UPDATE)
- **說明**：被邀請方同意時，系統會將該筆邀請的 `status` 從 `PENDING` 直接更新為 `ACCEPTED`，自此雙方結為好友。
- **對應 API**：`PUT /api/social/friends/request/:id/accept`
- **Go 實作 (GORM)**：
  ```go
  friend.Status = "ACCEPTED"
  database.DB.Save(&friend)
  ```
- **原生 SQL 語法**：
  ```sql
  UPDATE friendships SET status = 'ACCEPTED' WHERE friendship_id = 33;
  ```

### 12. 拒絕好友邀請 (UPDATE)
- **說明**：被邀請方不同意時，將狀態更新為 `DECLINED`，保留歷史紀錄。
- **對應 API**：`PUT /api/social/friends/request/:id/decline`
- **Go 實作 (GORM)**：
  ```go
  friend.Status = "DECLINED"
  database.DB.Save(&friend)
  ```
- **原生 SQL 語法**：
  ```sql
  UPDATE friendships SET status = 'DECLINED' WHERE friendship_id = 33;
  ```

### 13. 收回邀請或刪除好友 (DELETE)
- **說明**：主動收回尚未被處理的好友邀請，或是要解除已成立的好友關係時，系統會將該筆 `friendships` 紀錄實體刪除。
- **對應 API**：`DELETE /api/social/friends/request/:id`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Delete(&friend)
  ```
- **原生 SQL 語法**：
  ```sql
  DELETE FROM friendships WHERE friendship_id = 33;
  ```

### 14. 加入黑名單 (INSERT)
- **說明**：遭遇騷擾時將對方加入黑名單。這是一種軟封鎖，不會影響 `friendships` 表，但在後端傳送訊息等行為會被攔截。
- **對應 API**：`POST /api/social/blacklist`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Create(&blockRecord)
  ```
- **原生 SQL 語法**：
  ```sql
  INSERT INTO blacklists (blocker_id, blocked_id) VALUES (5, 10);
  ```

### 15. 移除黑名單 (DELETE)
- **說明**：解除封鎖，刪除對應的黑名單紀錄。
- **對應 API**：`DELETE /api/social/blacklist/:user_id`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Where("blocker_id = ? AND blocked_id = ?", userID, targetID).Delete(&models.Blocklist{})
  ```
- **原生 SQL 語法**：
  ```sql
  DELETE FROM blacklists WHERE blocker_id = 5 AND blocked_id = 10;
  ```

### 16. 發送對話訊息 (INSERT)
- **說明**：好友之間互相傳送私密訊息，建立紀錄時 `is_read` 預設為 `FALSE` (未讀)。
- **對應 API**：`POST /api/social/messages`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Create(&msg)
  ```
- **原生 SQL 語法**：
  ```sql
  INSERT INTO messages (sender_id, receiver_id, content) VALUES (5, 10, 'Hello!');
  ```

### 17. 標記訊息為已讀 (UPDATE)
- **說明**：當買家點開某位好友的聊天室時，系統會一口氣將對方傳給自己且還是未讀的訊息，全部更新為已讀狀態。
- **對應 API**：`GET /api/social/messages/{user_id}` (撈取同時順便標記)
- **Go 實作 (GORM)**：
  ```go
  database.DB.Model(&models.Message{}).Where("sender_id = ? AND receiver_id = ? AND is_read = ?", otherID, myID, false).Update("is_read", true)
  ```
- **原生 SQL 語法**：
  ```sql
  UPDATE messages SET is_read = true WHERE sender_id = 10 AND receiver_id = 5 AND is_read = false;
  ```

### 18. 發布筆記評論 (INSERT)
- **說明**：買家在購買筆記後發表心路歷程與評價，供其他買家參考。
- **對應 API**：`POST /api/social/notes/:id/reviews`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Create(&review)
  ```
- **原生 SQL 語法**：
  ```sql
  INSERT INTO reviews (note_id, user_id, attitude, content, status) 
  VALUES (42, 5, 'POSITIVE', '超好玩', 'VISIBLE');
  ```

### 19. 修改筆記評論 (UPDATE)
- **說明**：原本覺得好玩的筆記後來發現是個坑，買家可以自由修改評論內容與態度。
- **對應 API**：`PUT /api/social/reviews/:id` (雖然目前 API 清單未列出，但若實作則為 UPDATE)
- **Go 實作 (GORM)**：
  ```go
  database.DB.Save(&review)
  ```
- **原生 SQL 語法**：
  ```sql
  UPDATE reviews SET content = '玩久了有點膩', attitude = 'NEGATIVE' WHERE review_id = 128 AND user_id = 5;
  ```

### 20. 刪除筆記評論 (UPDATE 軟刪除)
- **說明**：買家刪除自己的評論時，為了資料完整性通常採用軟刪除，將 `status` 改為 `HIDDEN` 或 `DELETED`。
- **對應 API**：`DELETE /api/social/reviews/:id` (如果需要實作的話)
- **Go 實作 (GORM)**：
  ```go
  database.DB.Model(&review).Update("status", "HIDDEN")
  ```
- **原生 SQL 語法**：
  ```sql
  UPDATE reviews SET status = 'HIDDEN' WHERE review_id = 128 AND user_id = 5;
  ```

### 21. 回覆別人的評論 (INSERT)
- **說明**：俗稱樓中樓，允許買家對他人的評論進行附議或反駁。
- **對應 API**：`POST /api/social/reviews/:id/replies`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Create(&reply)
  ```
- **原生 SQL 語法**：
  ```sql
  INSERT INTO review_replies (review_id, user_id, content) VALUES (128, 10, '+1 認同');
  ```

### 22. 刪除回覆 (DELETE)
- **說明**：買家刪除自己留下的樓中樓回覆。
- **對應 API**：`DELETE /api/social/reviews/replies/:reply_id`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Delete(&reply)
  ```
- **原生 SQL 語法**：
  ```sql
  DELETE FROM review_replies WHERE review_reply_id = 56;
  ```

---

## 模組四：賣家功能 (Seller)

### 23. 發行新筆記草稿 (INSERT)
- **說明**：賣家建立一款新筆記的基本資料，此時筆記會是 `DRAFT` (草稿) 狀態，暫時不會在商店中曝光。
- **對應 API**：`POST /api/seller/notes`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Create(&note)
  ```
- **原生 SQL 語法**：
  ```sql
  INSERT INTO notes (title, description, price, seller_id, status) VALUES ('新筆記', '...', 500, 5, 'DRAFT');
  ```

### 24. 正式上架或編輯筆記資訊 (UPDATE)
- **說明**：補齊資料並加上至少一個科目後，賣家將筆記正式公開 (`ACTIVE`)，或者日後更新售價等欄位。
- **對應 API**：`PUT /api/seller/notes/:id/publish` 與 `PUT /api/seller/notes/:id`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Save(&note)
  ```
- **原生 SQL 語法**：
  ```sql
  UPDATE notes SET status = 'ACTIVE' WHERE note_id = 42 AND seller_id = 5;
  UPDATE notes SET title = '新標題', price = 400 WHERE note_id = 42 AND seller_id = 5;
  ```

### 25. 下架自己的筆記 (UPDATE)
- **說明**：賣家決定不再販售該筆記，將狀態改為 `TAKEN_DOWN`。這不會影響已經買過筆記的買家。
- **對應 API**：`DELETE /api/seller/notes/:id`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Model(&note).Update("status", "TAKEN_DOWN")
  ```
- **原生 SQL 語法**：
  ```sql
  UPDATE notes SET status = 'TAKEN_DOWN' WHERE note_id = 42 AND seller_id = 5;
  ```

### 26. 上傳筆記圖片或檔案 (INSERT)
- **說明**：上傳圖片預覽或筆記檔案案時，在資料庫記錄檔案的虛擬路徑與類型。
- **對應 API**：`POST /api/seller/notes/:id/media`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Create(&media)
  ```
- **原生 SQL 語法**：
  ```sql
  INSERT INTO note_media (note_id, media_type, file_url) VALUES (42, 'media', '/media/images/123.jpg');
  ```

### 27. 刪除媒體檔案 (DELETE)
- **說明**：若是上傳錯檔案或圖片，可以針對該媒體紀錄進行刪除。
- **對應 API**：`DELETE /api/seller/notes/:id/media/:media_id`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Delete(&media)
  ```
- **原生 SQL 語法**：
  ```sql
  DELETE FROM note_media WHERE media_id = 77;
  ```

### 28. 為筆記新增科目 (INSERT)
- **說明**：為了精準觸及受眾，賣家將筆記與特定科目綁定，這是建立多對多關係的操作。
- **對應 API**：`POST /api/seller/notes/:id/tags`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Create(&noteTag)
  ```
- **原生 SQL 語法**：
  ```sql
  INSERT INTO note_tags (note_id, tag_id) VALUES (42, 3);
  ```

### 29. 建立全域新科目與移除筆記科目 (INSERT / DELETE)
- **說明**：如果系統目前的科目不夠用，賣家可以創造新科目；若標錯了，也能夠從 `note_tags` 中解綁。
- **對應 API**：`POST /api/seller/tags` 與 `DELETE /api/seller/notes/:id/tags/:tag_id`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Create(&newTag)
  database.DB.Where("note_id = ? AND tag_id = ?", noteID, tagID).Delete(&models.NoteTag{})
  ```
- **原生 SQL 語法**：
  ```sql
  INSERT INTO tags (tag_name) VALUES ('MOBA');
  DELETE FROM note_tags WHERE note_id = 42 AND tag_id = 3;
  ```

---

## 模組五：管理員與客服 (Admin & CSR)

### 30. 管理員停權/切換買家身分 (UPDATE)
- **說明**：系統最高管理員對惡意買家進行停權處分，或是拔擢某位買家成為客服人員 (`CSR`)。
- **對應 API**：`PUT /api/admin/users/:id/suspend` 與 `PUT /api/admin/users/:id/role`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Save(&user)
  ```
- **原生 SQL 語法**：
  ```sql
  UPDATE users SET permission = 'DEACTIVE' WHERE user_id = 10;
  UPDATE users SET role = 'CSR' WHERE user_id = 10;
  ```

### 31. 管理員強制刪除帳號 (UPDATE 軟刪除 + 級聯操作)
- **說明**：管理員對帳號施以終極極刑。不僅將帳號改為 `DELETED`，還會自動觸發下述的強制下架與撤銷買家授權。
- **對應 API**：`DELETE /api/admin/users/:id`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Save(&user) // Permission = 'DELETED'
  ```
- **原生 SQL 語法**：
  ```sql
  UPDATE users SET permission = 'DELETED' WHERE user_id = 10;
  ```

### 32. 管理員強制下架筆記與終極撤銷 (UPDATE)
- **說明**：當筆記嚴重違規時，管理員將之強制下架。與賣家自主下架不同的是，管理員的鐵腕會直接追殺到已購買的買家，一併將他們的授權改為 `REVOKED`。
- **對應 API**：`DELETE /api/admin/notes/:id`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Model(&note).Update("status", "TAKEN_DOWN")
  database.DB.Model(&models.NoteLicense{}).Where("note_id = ?", noteID).Update("status", "REVOKED")
  ```
- **原生 SQL 語法**：
  ```sql
  UPDATE notes SET status = 'TAKEN_DOWN' WHERE note_id = 42;
  UPDATE note_licenses SET status = 'REVOKED' WHERE note_id = 42;
  ```

### 33. 客服同意或拒絕退款單 (UPDATE)
- **說明**：客服完成查核後，將退款單狀態改為 `APPROVED` 或 `REJECTED`。若同意退款，則必須將買家的筆記授權廢除 (`REVOKED`)。
- **對應 API**：`PUT /api/csr/refunds/:id`
- **Go 實作 (GORM)**：
  ```go
  database.DB.Save(&request)
  database.DB.Model(&license).Update("status", "REVOKED") // 若核准
  ```
- **原生 SQL 語法**：
  ```sql
  UPDATE refund_requests SET status = 'APPROVED', resolved_at = NOW() WHERE refund_id = 88;
  UPDATE note_licenses SET status = 'REVOKED' WHERE transaction_item_id = 105; 
  ```
