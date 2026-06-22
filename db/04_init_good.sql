-- ==========================================
-- 10. 新增三個筆記與其對應的 Media
-- ==========================================
WITH inserted_notes AS (
    INSERT INTO notes (seller_id, title, description, price, overall_rating, status)
    VALUES 
        (2, '物理搖擺實驗數據分析筆記', '包含單擺與簡諧運動的實驗數據與公式分析', 10.00, 4.5, 'ACTIVE'),
        (2, '流體力學水上飛行動力分析', '詳細探討水上飛行的流體力學與受力平衡原理', 20.00, 4.8, 'ACTIVE'),
        (2, '機翼空氣動力學模擬筆記', '整理機翼受力、升力係數與阻力係數的模擬與重點', 30.00, 4.2, 'ACTIVE')
    RETURNING note_id, title
)
INSERT INTO note_media (note_id, file_url, media_type)
SELECT 
    note_id,
    CASE 
        WHEN title = '物理搖擺實驗數據分析筆記' THEN '/media/images/' || note_id || '/shake.gif'
        WHEN title = '流體力學水上飛行動力分析' THEN '/media/images/' || note_id || '/water_fly.gif'
        WHEN title = '機翼空氣動力學模擬筆記' THEN '/media/images/' || note_id || '/wing.gif'
    END,
    'media'
FROM inserted_notes;

-- ==========================================
-- 11. Assign proper tags to the 3 good notes
--     Each note: 1 SEMESTER + 1 SUBJECT + 1 DEPARTMENT + 1 COURSE_TYPE + 1 TEACHER
-- ==========================================

-- 物理搖擺實驗數據分析筆記: 114-1, 應用數學系, 必修
INSERT INTO note_tags (note_id, tag_id)
SELECT n.note_id, t.tag_id FROM notes n, tags t
WHERE n.title = '物理搖擺實驗數據分析筆記' AND t.tag_name = '114-1' AND t.tag_type = 'SEMESTER'
ON CONFLICT DO NOTHING;

INSERT INTO note_tags (note_id, tag_id)
SELECT n.note_id, t.tag_id FROM notes n, tags t
WHERE n.title = '物理搖擺實驗數據分析筆記' AND t.tag_name = '應用數學系' AND t.tag_type = 'DEPARTMENT'
ON CONFLICT DO NOTHING;

INSERT INTO note_tags (note_id, tag_id)
SELECT n.note_id, t.tag_id FROM notes n, tags t
WHERE n.title = '物理搖擺實驗數據分析筆記' AND t.tag_name = '必修' AND t.tag_type = 'COURSE_TYPE'
ON CONFLICT DO NOTHING;

INSERT INTO note_tags (note_id, tag_id)
SELECT n.note_id, t.tag_id FROM notes n, tags t
WHERE n.title = '物理搖擺實驗數據分析筆記' AND t.tag_name = '蔡國裕' AND t.tag_type = 'TEACHER'
ON CONFLICT DO NOTHING;

-- 流體力學水上飛行動力分析: 113-2, 電機工程學系, 選修
INSERT INTO note_tags (note_id, tag_id)
SELECT n.note_id, t.tag_id FROM notes n, tags t
WHERE n.title = '流體力學水上飛行動力分析' AND t.tag_name = '113-2' AND t.tag_type = 'SEMESTER'
ON CONFLICT DO NOTHING;

INSERT INTO note_tags (note_id, tag_id)
SELECT n.note_id, t.tag_id FROM notes n, tags t
WHERE n.title = '流體力學水上飛行動力分析' AND t.tag_name = '電機工程學系' AND t.tag_type = 'DEPARTMENT'
ON CONFLICT DO NOTHING;

INSERT INTO note_tags (note_id, tag_id)
SELECT n.note_id, t.tag_id FROM notes n, tags t
WHERE n.title = '流體力學水上飛行動力分析' AND t.tag_name = '選修' AND t.tag_type = 'COURSE_TYPE'
ON CONFLICT DO NOTHING;

INSERT INTO note_tags (note_id, tag_id)
SELECT n.note_id, t.tag_id FROM notes n, tags t
WHERE n.title = '流體力學水上飛行動力分析' AND t.tag_name = '蔡國裕' AND t.tag_type = 'TEACHER'
ON CONFLICT DO NOTHING;

-- 機翼空氣動力學模擬筆記: 114-2, 資訊工程學系, 選修
INSERT INTO note_tags (note_id, tag_id)
SELECT n.note_id, t.tag_id FROM notes n, tags t
WHERE n.title = '機翼空氣動力學模擬筆記' AND t.tag_name = '114-2' AND t.tag_type = 'SEMESTER'
ON CONFLICT DO NOTHING;

INSERT INTO note_tags (note_id, tag_id)
SELECT n.note_id, t.tag_id FROM notes n, tags t
WHERE n.title = '機翼空氣動力學模擬筆記' AND t.tag_name = '資訊工程學系' AND t.tag_type = 'DEPARTMENT'
ON CONFLICT DO NOTHING;

INSERT INTO note_tags (note_id, tag_id)
SELECT n.note_id, t.tag_id FROM notes n, tags t
WHERE n.title = '機翼空氣動力學模擬筆記' AND t.tag_name = '選修' AND t.tag_type = 'COURSE_TYPE'
ON CONFLICT DO NOTHING;

INSERT INTO note_tags (note_id, tag_id)
SELECT n.note_id, t.tag_id FROM notes n, tags t
WHERE n.title = '機翼空氣動力學模擬筆記' AND t.tag_name = '蔡國裕' AND t.tag_type = 'TEACHER'
ON CONFLICT DO NOTHING;

-- ==========================================
-- 12. Randomize prices for existing test data (between 31 and 400)
-- ==========================================
UPDATE notes
SET price = FLOOR(RANDOM() * 370 + 31)
WHERE title NOT IN ('物理搖擺實驗數據分析筆記', '流體力學水上飛行動力分析', '機翼空氣動力學模擬筆記');

-- ==========================================
-- 13. Safety net: ensure ALL notes have exactly 1 SEMESTER tag
--     (fills any note that somehow has 0 semester tags)
-- ==========================================
INSERT INTO note_tags (note_id, tag_id)
SELECT n.note_id, (
    SELECT tag_id FROM tags WHERE tag_type = 'SEMESTER' ORDER BY RANDOM() LIMIT 1
)
FROM notes n
WHERE NOT EXISTS (
    SELECT 1 FROM note_tags nt
    JOIN tags t ON nt.tag_id = t.tag_id
    WHERE nt.note_id = n.note_id AND t.tag_type = 'SEMESTER'
)
ON CONFLICT DO NOTHING;
