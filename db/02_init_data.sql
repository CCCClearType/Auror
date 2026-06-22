-- ==========================================
-- 0. 清除舊資料並重置 ID (避免重複插入與 ID 錯亂)
-- ==========================================
TRUNCATE TABLE users, notes, tags, note_tags, note_media, transactions, transaction_items, shopping_carts, note_licenses, wish_lists, refund_requests, reviews, review_replies, friendships, messages, blacklists RESTART IDENTITY CASCADE;

-- ==========================================
-- 1. 寫入 users 資料 (5筆)
-- ==========================================
INSERT INTO users (username, email, password_hash, last_visit_ip, role, status, permission) VALUES 
('AdminMaster', 'admin@aurorvapor.com', 'hashed_pwd_001', '192.168.1.1', 'ADMIN', 'ONLINE', 'ACTIVE'),
('StudioAurora', 'dev@studioaurora.com', 'hashed_pwd_002', '10.0.0.15', 'SELLER', 'OFFLINE', 'ACTIVE'),
('SupportAlice', 'alice@support.com', 'hashed_pwd_003', '172.16.0.25', 'CSR', 'ONLINE', 'ACTIVE'),
('PlayerOne', 'player1@gmail.com', 'hashed_pwd_004', '114.34.56.78', 'USERS', 'OFFLINE', 'DEACTIVE'),
('Guest999', 'guest999@tempmail.com', 'hashed_pwd_005', '2001:0db8:85a3::7334', 'NULL', 'OFFLINE', 'DELETED');

-- ==========================================
-- 2. 寫入 notes 資料 (4筆)
-- ==========================================
INSERT INTO notes (seller_id, title, description, price, overall_rating, status) VALUES
(1, '資料結構期末考精華筆記', '涵蓋樹、圖、排序等核心資料結構概念與考題精選。', 1200.00, 4.5, 'ACTIVE'),
(1, '演算法章節重點總整理', '分析動態規劃、貪婪演算法與分治法等必考重點。', 150.00, 4.8, 'ACTIVE'),
(2, '材料與生活通識課必讀筆記', '整理材料性質與日常生活應用，輕鬆取得高分。', 850.50, 3.9, 'ACTIVE'),
(2, '微積分甲高分全攻略', '包含極限、導數、積分及多元微積分推導與習題詳解。', 300.00, 4.2, 'ACTIVE');

-- ==========================================
-- 3. 寫入 tags 與 note_tags 資料
-- ==========================================
INSERT INTO tags (tag_name, tag_type) VALUES
('資料結構', 'SUBJECT'), ('演算法', 'SUBJECT'), ('材料與生活', 'SUBJECT'), ('微積分', 'SUBJECT'),
('113-1', 'SEMESTER'), ('112-2', 'SEMESTER'), ('112-1', 'SEMESTER');

INSERT INTO note_tags (note_id, tag_id) VALUES
(1, 1), (1, 2), (2, 1), (3, 3), (4, 4), -- Subject tags
(1, 5), (2, 5), (3, 6), (4, 7);         -- Semester tags

-- ==========================================
-- 4. 寫入 note_media 資料
--    圖片/影片 (media): /media/images/{note_id}/{sha256}.{ext}
--    筆記檔案 (note_file): /downloads/{note_id}/{original_name}
-- ==========================================
INSERT INTO note_media (note_id, file_url, media_type) VALUES
(1, '/media/images/1/63cb3d94925658f69d65f10a8a529599b42c0faaf97559f3212acc085c5d4da7.jpg', 'media'),
(1, '/media/images/1/2899c89bd64ddca70f023a2bf9d4a0c77897dc990fe260aabe2e3814259559b7.jpg', 'media'),
(1, '/media/images/1/acaba9e0fae5fa9d72cfe96b930ccdc4d447cce078bd6eff0ee73b869019530c.png', 'media'),
(1, '/media/images/1/cec7e79ee9686b5ca9fa4112ed16db99cb4c5c3f55440a065935a1e7ff171582.png', 'media'),
(1, '/media/images/1/cac363c57e461ca353b5b61a8dc7a865a989f1ee606c530ef4f4c76bbdea7142.png', 'media'),

(2, '/media/images/2/hot dog.gif', 'media'),
(2, '/media/images/2/86.jpg', 'media'),

(3, '/media/images/3/acaba9e0fae5fa9d72cfe96b930ccdc4d447cce078bd6eff0ee73b869019530c.png', 'media'),
(3, '/media/images/3/acaba9e0fae5fa9d72cfe96b930ccdc4d447cce078bd6eff0ee73b869019530c.png', 'media'),
(3, '/media/images/3/cec7e79ee9686b5ca9fa4112ed16db99cb4c5c3f55440a065935a1e7ff171582.png', 'media'),

(1, '/downloads/1/data-structures-notes.txt', 'note_file'),
(3, '/downloads/3/bathroom.png', 'note_file'),
(4, '/downloads/4/calculus-guide.txt', 'note_file');

-- ==========================================
-- 5. 寫入 transactions 與 transaction_items 資料
-- ==========================================
INSERT INTO transactions (user_id, total_amount, receipt_number) VALUES
(1, 1350.00, 'REC-20260605-0001'),
(2, 300.00, 'REC-20260605-0002'),
(1, 850.50, 'REC-20260606-0003'),
(2, 1200.00, 'REC-20260607-0004');

INSERT INTO transaction_items (transaction_id, note_id, purchase_price) VALUES
(1, 1, 1200.00), -- item_id 1
(1, 2, 150.00),  -- item_id 2
(2, 4, 300.00),  -- item_id 3
(3, 3, 850.50),  -- item_id 4
(4, 1, 1200.00); -- item_id 5

-- ==========================================
-- 6. 寫入 shopping_carts 資料 (4筆)
-- ==========================================
INSERT INTO shopping_carts (user_id, note_id) VALUES
(1, 3), (1, 4), (2, 1), (2, 2);

-- ==========================================
-- 7. 寫入 note_licenses 與 wish_lists 資料
-- ==========================================
INSERT INTO note_licenses (note_id, user_id, transaction_item_id, status) VALUES
(1, 1, 1, 'ACTIVE'), 
(2, 1, 2, 'REVOKED'), 
(4, 2, 3, 'ACTIVE'), 
(3, 1, 4, 'REVOKED');

INSERT INTO wish_lists (user_id, note_id) VALUES
(1, 3), (1, 4), (2, 1), (2, 2);

-- ==========================================
-- 8. 寫入 refund_requests 退款資料 
-- ==========================================
INSERT INTO refund_requests (buyer_id, transaction_item_id, handled_by, reason, reject_reason, resolved_at, status) VALUES 
(1, 1, NULL, '買錯筆記了，內容跟我們班教授教的不一樣', NULL, NULL, 'PENDING'),
(1, 2, 3, '筆記檔案損毀，無法正常開啟閱讀', NULL, '2026-06-05 14:30:00', 'APPROVED'),
(2, 3, 3, '覺得內容不太充實，想申請退款', '您已下載此筆記，不符合平台的退款政策。', '2026-06-06 09:15:00', 'REJECTED'),
(1, 4, 1, '筆記內容嚴重錯誤，要求處理', NULL, '2026-06-06 10:00:00', 'APPROVED');

-- ==========================================
-- 9. 寫入社交系統資料 (reviews, replies, friends, msgs, blacklists)
-- ==========================================
INSERT INTO reviews (note_id, user_id, content, attitude) VALUES
(1, 1, '神作！排版超讚，內容非常豐富。', 'POSITIVE'),
(3, 2, '排版有點混亂，希望能盡快更新。', 'NEGATIVE'),
(1, 3, '雖然有少許錯字，但不影響整體極佳的體驗。', 'POSITIVE'),
(4, 1, '微積分整理得太療癒了，可以看一整天。', 'POSITIVE');

-- 為了避免時序錯誤，將回覆拆分為獨立語句依序寫入
INSERT INTO review_replies (review_id, user_id, parent_reply_id, content) VALUES
(2, 1, NULL, '真的！特別是圖形演算法那段排版卡卡的。');

INSERT INTO review_replies (review_id, user_id, parent_reply_id, content) VALUES
(2, 2, 1, '原來大家都有這個問題，我以為是我看不懂。');

INSERT INTO review_replies (review_id, user_id, parent_reply_id, content) VALUES
(1, 4, NULL, '完全同意，年度最佳筆記無誤！');

INSERT INTO review_replies (review_id, user_id, parent_reply_id, content) VALUES
(1, 1, 3, '謝謝版主的認同！');

INSERT INTO friendships (sender_id, receiver_id, status) VALUES
(1, 2, 'ACCEPTED'), (3, 1, 'ACCEPTED'), (2, 4, 'PENDING'), (4, 1, 'DECLINED');

INSERT INTO messages (sender_id, receiver_id, content, is_read) VALUES
(1, 2, '今晚要一起討論期末考進度嗎？', TRUE),
(2, 1, '好啊！我大概晚上八點去圖書館。', FALSE),
(3, 1, '嗨，你的退款申請我們收到了。', TRUE),
(1, 3, '好的，麻煩您了。', FALSE);

INSERT INTO blacklists (blocker_id, blocked_id) VALUES
(1, 4), (2, 3), (4, 2), (3, 4);
