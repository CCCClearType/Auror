import random

subjects = ['資料結構', '演算法', '計算機結構', '作業系統', '計算機網路', '資料庫系統', '軟體工程', '人工智慧', '機器學習', '密碼學']
semesters = ['113-1', '112-2', '112-1', '114-1', '114-2']
departments = ['資訊工程學系', '電機工程學系', '企業管理學系', '應用數學系', '通識教育中心']
course_types = ['必修', '選修', '通識']
teachers = ['林智維', '張家銘', '李明哲', '王大明', '陳德生', '吳宗憲', '黃健明', '薛念林']

with open('db/03_init_super_data.sql', 'r', encoding='utf-8', errors='ignore') as f:
    lines = f.readlines()

new_lines = []
skip = False
note_ids = set()

for line in lines:
    if line.startswith("-- PATCH:"):
        skip = True
    if "INSERT INTO tags (tag_id, tag_name) VALUES (10" in line:
        continue # remove corrupted tags
    if "INSERT INTO note_tags (note_id, tag_id) VALUES (" in line and ("10" in line or "11" in line):
        continue # remove original mappings (which used tags 100-109)
    
    if "INSERT INTO notes (note_id" in line:
        parts = line.split("VALUES (")
        if len(parts) > 1:
            nid = parts[1].split(",")[0].strip()
            if nid.isdigit():
                note_ids.add(int(nid))
    
    if not skip:
        new_lines.append(line)

new_lines.append("\n-- PATCH: Automatically assign tags\n")
for idx, s in enumerate(subjects):
    new_lines.append(f"INSERT INTO tags (tag_id, tag_name, tag_type) VALUES ({100+idx}, '{s}', 'SUBJECT') ON CONFLICT DO NOTHING;\n")

# Make sure the dynamically added tags exist
for t in semesters:
    new_lines.append(f"INSERT INTO tags (tag_name, tag_type) VALUES ('{t}', 'SEMESTER') ON CONFLICT DO NOTHING;\n")
for t in departments:
    new_lines.append(f"INSERT INTO tags (tag_name, tag_type) VALUES ('{t}', 'DEPARTMENT') ON CONFLICT DO NOTHING;\n")
for t in course_types:
    new_lines.append(f"INSERT INTO tags (tag_name, tag_type) VALUES ('{t}', 'COURSE_TYPE') ON CONFLICT DO NOTHING;\n")
for t in teachers:
    new_lines.append(f"INSERT INTO tags (tag_name, tag_type) VALUES ('{t}', 'TEACHER') ON CONFLICT DO NOTHING;\n")

for nid in sorted(list(note_ids)):
    s = random.choice(semesters)
    sub = random.choice(subjects)
    d = random.choice(departments)
    c = random.choice(course_types)
    new_lines.append(f"INSERT INTO note_tags (note_id, tag_id) SELECT {nid}, tag_id FROM tags WHERE tag_name = '{s}' ON CONFLICT DO NOTHING;\n")
    new_lines.append(f"INSERT INTO note_tags (note_id, tag_id) SELECT {nid}, tag_id FROM tags WHERE tag_name = '{sub}' ON CONFLICT DO NOTHING;\n")
    new_lines.append(f"INSERT INTO note_tags (note_id, tag_id) SELECT {nid}, tag_id FROM tags WHERE tag_name = '{d}' ON CONFLICT DO NOTHING;\n")
    new_lines.append(f"INSERT INTO note_tags (note_id, tag_id) SELECT {nid}, tag_id FROM tags WHERE tag_name = '{c}' ON CONFLICT DO NOTHING;\n")
    if random.random() > 0.3:
        t = random.choice(teachers)
        new_lines.append(f"INSERT INTO note_tags (note_id, tag_id) SELECT {nid}, tag_id FROM tags WHERE tag_name = '{t}' ON CONFLICT DO NOTHING;\n")

with open('db/03_init_super_data.sql', 'w', encoding='utf-8') as f:
    f.writelines(new_lines)
