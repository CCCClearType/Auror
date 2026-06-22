import random

# Get the list of all note_ids in 03_init_super_data.sql by searching for INSERT INTO notes
note_ids = []
with open('db/03_init_super_data.sql', 'r', encoding='utf-8') as f:
    for line in f:
        if "INSERT INTO notes (note_id" in line:
            parts = line.split("VALUES (")
            if len(parts) > 1:
                nid = parts[1].split(",")[0].strip()
                if nid.isdigit():
                    note_ids.append(int(nid))

semesters = [(200, '113-1', 'SEMESTER'), (201, '112-2', 'SEMESTER'), (202, '112-1', 'SEMESTER'), (203, '114-1', 'SEMESTER'), (204, '114-2', 'SEMESTER')]
departments = [(210, '資訊工程學系', 'DEPARTMENT'), (211, '電機工程學系', 'DEPARTMENT'), (212, '企業管理學系', 'DEPARTMENT'), (213, '通識教育中心', 'DEPARTMENT')]
course_types = [(220, '必修', 'COURSE_TYPE'), (221, '選修', 'COURSE_TYPE'), (222, '通識', 'COURSE_TYPE')]
teachers = [(230, '林哲維', 'TEACHER'), (231, '陳德生', 'TEACHER'), (232, '張真誠', 'TEACHER'), (233, '薛念林', 'TEACHER')]

lines = [
    "\n-- ===========================================================================",
    "-- PATCH: Automatically assign SEMESTER, DEPARTMENT, COURSE_TYPE, TEACHER tags",
    "-- ===========================================================================",
    "UPDATE tags SET tag_type = 'SUBJECT' WHERE tag_id BETWEEN 100 AND 109;"
]

all_new_tags = semesters + departments + course_types + teachers
for tid, tname, ttype in all_new_tags:
    lines.append(f"INSERT INTO tags (tag_id, tag_name, tag_type) VALUES ({tid}, '{tname}', '{ttype}') ON CONFLICT DO NOTHING;")

for nid in note_ids:
    s = random.choice(semesters)[0]
    d = random.choice(departments)[0]
    c = random.choice(course_types)[0]
    lines.append(f"INSERT INTO note_tags (note_id, tag_id) VALUES ({nid}, {s}) ON CONFLICT DO NOTHING;")
    lines.append(f"INSERT INTO note_tags (note_id, tag_id) VALUES ({nid}, {d}) ON CONFLICT DO NOTHING;")
    lines.append(f"INSERT INTO note_tags (note_id, tag_id) VALUES ({nid}, {c}) ON CONFLICT DO NOTHING;")
    if random.random() > 0.3:
        t = random.choice(teachers)[0]
        lines.append(f"INSERT INTO note_tags (note_id, tag_id) VALUES ({nid}, {t}) ON CONFLICT DO NOTHING;")

with open('db/03_init_super_data.sql', 'a', encoding='utf-8') as f:
    f.write('\n'.join(lines) + '\n')

print(f"Patched 03_init_super_data.sql with {len(all_new_tags)} new tags and assigned to {len(note_ids)} notes.")
