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
