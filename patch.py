import random

note_ids = []
with open('db/03_init_super_data.sql', 'r', encoding='utf-8') as f:
    for line in f:
        if "INSERT INTO notes (note_id" in line:
            parts = line.split("VALUES (")
            if len(parts) > 1:
                nid = parts[1].split(",")[0].strip()
                if nid.isdigit():
                    note_ids.append(int(nid))

semesters = [('113-1', 'SEMESTER'), ('112-2', 'SEMESTER'), ('112-1', 'SEMESTER'), ('114-1', 'SEMESTER'), ('114-2', 'SEMESTER')]
departments = [('資訊工程學系', 'DEPARTMENT'), ('電機工程學系', 'DEPARTMENT'), ('企業管理學系', 'DEPARTMENT'), ('通識教育中心', 'DEPARTMENT')]
course_types = [('必修', 'COURSE_TYPE'), ('選修', 'COURSE_TYPE'), ('通識', 'COURSE_TYPE')]
teachers = [('林哲維', 'TEACHER'), ('陳德生', 'TEACHER'), ('張真誠', 'TEACHER'), ('薛念林', 'TEACHER')]

lines = [
    "\n-- ===========================================================================",
    "-- PATCH: Automatically assign SEMESTER, DEPARTMENT, COURSE_TYPE, TEACHER tags",
    "-- ===========================================================================",
    "UPDATE tags SET tag_type = 'SUBJECT' WHERE tag_id BETWEEN 100 AND 109;"
]

all_new_tags = semesters + departments + course_types + teachers
for tname, ttype in all_new_tags:
    lines.append(f"INSERT INTO tags (tag_name, tag_type) VALUES ('{tname}', '{ttype}') ON CONFLICT DO NOTHING;")

for nid in note_ids:
    s = random.choice(semesters)[0]
    d = random.choice(departments)[0]
    c = random.choice(course_types)[0]
    lines.append(f"INSERT INTO note_tags (note_id, tag_id) SELECT {nid}, tag_id FROM tags WHERE tag_name = '{s}' ON CONFLICT DO NOTHING;")
    lines.append(f"INSERT INTO note_tags (note_id, tag_id) SELECT {nid}, tag_id FROM tags WHERE tag_name = '{d}' ON CONFLICT DO NOTHING;")
    lines.append(f"INSERT INTO note_tags (note_id, tag_id) SELECT {nid}, tag_id FROM tags WHERE tag_name = '{c}' ON CONFLICT DO NOTHING;")
    if random.random() > 0.3:
        t = random.choice(teachers)[0]
        lines.append(f"INSERT INTO note_tags (note_id, tag_id) SELECT {nid}, tag_id FROM tags WHERE tag_name = '{t}' ON CONFLICT DO NOTHING;")

with open('db/03_init_super_data.sql', 'a', encoding='utf-8') as f:
    f.write('\n'.join(lines) + '\n')

print(f"Patched 03_init_super_data.sql with {len(all_new_tags)} new tags and assigned to {len(note_ids)} notes using dynamic tag_id.")
