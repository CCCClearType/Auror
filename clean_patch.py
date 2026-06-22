"""
clean_patch.py  –  Rebuild the tag-assignment block in 03_init_super_data.sql
Rules:
  • Every super-data note gets EXACTLY 1 SEMESTER, 1 SUBJECT, 1 DEPARTMENT,
    1 COURSE_TYPE, and (70 % chance) 1 TEACHER.
  • The old hardcoded "Insert Semester Note Tags" block that gave every note
    tag_id 5 (= 113-1) is deleted.
  • The old "-- PATCH:" block is deleted and replaced by a clean one.
  • Semesters are the fixed set 108-1 … 115-2 (system-managed, 16 values).
"""

import random, re

# ── Reference data ──────────────────────────────────────────────
SEMESTERS = [
    '108-1', '108-2',
    '109-1', '109-2',
    '110-1', '110-2',
    '111-1', '111-2',
    '112-1', '112-2',
    '113-1', '113-2',
    '114-1', '114-2',
    '115-1', '115-2',
]

SUBJECTS = [
    '資料結構', '演算法', '計算機結構', '作業系統', '計算機網路',
    '資料庫系統', '軟體工程', '人工智慧', '機器學習', '密碼學',
]

DEPARTMENTS = [
    '資訊工程學系', '電機工程學系', '企業管理學系', '應用數學系', '通識教育中心',
]

COURSE_TYPES = ['必修', '選修', '通識']

TEACHERS = [
    '林智維', '張家銘', '李明哲', '王大明',
    '陳德生', '吳宗憲', '黃健明', '薛念林',
]

# ── Read ────────────────────────────────────────────────────────
with open('db/03_init_super_data.sql', 'r', encoding='utf-8') as f:
    content = f.read()

# ── Collect note IDs ────────────────────────────────────────────
note_ids = set()
for m in re.finditer(r"INSERT INTO notes \(note_id.*?VALUES\s*\((\d+),", content):
    note_ids.add(int(m.group(1)))

# ── Remove old semester-tag block ───────────────────────────────
# This block starts with "-- Insert Semester Note Tags" and ends at the
# closing ";", possibly spanning many lines.
content = re.sub(
    r'-- Insert Semester Note Tags for Super Data\s*\n'
    r'INSERT INTO note_tags \(note_id, tag_id\) VALUES\s*\n'
    r'(?:\(\d+,\s*\d+\)[,;]\s*\n?)+',
    '',
    content,
    flags=re.DOTALL,
)

# ── Remove old PATCH block (everything from "-- PATCH:" to EOF) ─
patch_pos = content.find('-- PATCH:')
if patch_pos != -1:
    content = content[:patch_pos]

# ── Build the new PATCH ─────────────────────────────────────────
lines = ['\n-- PATCH: System-managed semester tags + per-note tag assignments\n']

# 1) Ensure all 16 semesters exist
for s in SEMESTERS:
    lines.append(
        f"INSERT INTO tags (tag_name, tag_type) "
        f"VALUES ('{s}', 'SEMESTER') ON CONFLICT DO NOTHING;\n"
    )

# 2) Ensure subjects 100-109 exist
for idx, s in enumerate(SUBJECTS):
    lines.append(
        f"INSERT INTO tags (tag_id, tag_name, tag_type) "
        f"VALUES ({100 + idx}, '{s}', 'SUBJECT') ON CONFLICT DO NOTHING;\n"
    )

# 3) Ensure departments exist
for d in DEPARTMENTS:
    lines.append(
        f"INSERT INTO tags (tag_name, tag_type) "
        f"VALUES ('{d}', 'DEPARTMENT') ON CONFLICT DO NOTHING;\n"
    )

# 4) Ensure course types exist
for c in COURSE_TYPES:
    lines.append(
        f"INSERT INTO tags (tag_name, tag_type) "
        f"VALUES ('{c}', 'COURSE_TYPE') ON CONFLICT DO NOTHING;\n"
    )

# 5) Ensure teachers exist
for t in TEACHERS:
    lines.append(
        f"INSERT INTO tags (tag_name, tag_type) "
        f"VALUES ('{t}', 'TEACHER') ON CONFLICT DO NOTHING;\n"
    )

lines.append('\n')

# 6) For each note: exactly 1 semester + 1 subject + 1 dept + 1 course_type + maybe teacher
# First, delete any stale SEMESTER tag links for these notes so we get exactly 1
lines.append("-- Remove any pre-existing SEMESTER tags for super data notes\n")
lines.append(
    "DELETE FROM note_tags WHERE note_id >= 101 AND tag_id IN "
    "(SELECT tag_id FROM tags WHERE tag_type = 'SEMESTER');\n\n"
)

for nid in sorted(note_ids):
    sem   = random.choice(SEMESTERS)
    subj  = random.choice(SUBJECTS)
    dept  = random.choice(DEPARTMENTS)
    ctype = random.choice(COURSE_TYPES)

    lines.append(
        f"INSERT INTO note_tags (note_id, tag_id) "
        f"SELECT {nid}, tag_id FROM tags WHERE tag_name = '{sem}' "
        f"AND tag_type = 'SEMESTER' ON CONFLICT DO NOTHING;\n"
    )
    lines.append(
        f"INSERT INTO note_tags (note_id, tag_id) "
        f"SELECT {nid}, tag_id FROM tags WHERE tag_name = '{subj}' "
        f"AND tag_type = 'SUBJECT' ON CONFLICT DO NOTHING;\n"
    )
    lines.append(
        f"INSERT INTO note_tags (note_id, tag_id) "
        f"SELECT {nid}, tag_id FROM tags WHERE tag_name = '{dept}' "
        f"AND tag_type = 'DEPARTMENT' ON CONFLICT DO NOTHING;\n"
    )
    lines.append(
        f"INSERT INTO note_tags (note_id, tag_id) "
        f"SELECT {nid}, tag_id FROM tags WHERE tag_name = '{ctype}' "
        f"AND tag_type = 'COURSE_TYPE' ON CONFLICT DO NOTHING;\n"
    )
    if random.random() > 0.3:
        teacher = random.choice(TEACHERS)
        lines.append(
            f"INSERT INTO note_tags (note_id, tag_id) "
            f"SELECT {nid}, tag_id FROM tags WHERE tag_name = '{teacher}' "
            f"AND tag_type = 'TEACHER' ON CONFLICT DO NOTHING;\n"
        )

# ── Write ───────────────────────────────────────────────────────
with open('db/03_init_super_data.sql', 'w', encoding='utf-8') as f:
    f.write(content)
    f.writelines(lines)

print(f"Done. Patched {len(note_ids)} notes with exactly 1 semester each.")
