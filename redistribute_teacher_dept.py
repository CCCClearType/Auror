import re, random
import subprocess
import json

random.seed(42)

def get_tags_from_db(tag_type):
    cmd = f'docker compose exec -T db psql -U admin -d auror_vapor -t -c "SELECT tag_name FROM tags WHERE tag_type = \'{tag_type}\';"'
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True, encoding='utf-8')
    tags = [line.strip() for line in result.stdout.split('\n') if line.strip()]
    return tags

all_teachers = get_tags_from_db('TEACHER')
all_depts = get_tags_from_db('DEPARTMENT')

print(f"Found {len(all_teachers)} teachers and {len(all_depts)} departments.")

with open('db/03_init_super_data.sql', 'r', encoding='utf-8') as f:
    lines = f.read().split('\n')

old_teachers = ['薛念林', '李明哲', '王大明', '張家銘', '陳德生', '吳宗憲', '林智維', '黃健明']
zero_teachers = [t for t in all_teachers if t not in old_teachers]

old_depts = ['通識教育中心', '資訊工程學系', '應用數學系', '電機工程學系', '企業管理學系']
zero_depts = [d for d in all_depts if d not in old_depts]

def redistribute_type(lines, tag_type, all_items, old_items, zero_items, old_keep_min, old_keep_max, new_assign_min, new_assign_max):
    pattern = re.compile(
        f"INSERT INTO note_tags \\(note_id, tag_id\\) SELECT (\\d+), tag_id FROM tags WHERE tag_name = '([^']+)' AND tag_type = '{tag_type}'"
    )
    
    note_lines = {}
    for i, line in enumerate(lines):
        m = pattern.search(line)
        if m:
            note_id = int(m.group(1))
            item = m.group(2)
            if note_id not in note_lines:
                note_lines[note_id] = []
            note_lines[note_id].append((i, item))
            
    all_assignments = []
    for note_id, entries in sorted(note_lines.items()):
        for line_idx, item in entries:
            all_assignments.append((line_idx, note_id, item))
            
    random.shuffle(all_assignments)
    
    old_keep_count = {s: 0 for s in old_items}
    old_keep_max_dict = {s: random.randint(old_keep_min, old_keep_max) for s in old_items}
    
    new_assign_count = {s: 0 for s in zero_items}
    new_assign_max_dict = {s: random.randint(new_assign_min, new_assign_max) for s in zero_items}
    
    shuffled_new = list(zero_items)
    random.shuffle(shuffled_new)
    new_idx = 0
    
    replacements = {}
    
    for line_idx, note_id, old_item in all_assignments:
        if old_item in old_keep_count and old_keep_count[old_item] < old_keep_max_dict.get(old_item, 0):
            old_keep_count[old_item] += 1
            continue
            
        found = False
        attempts = 0
        while attempts < len(shuffled_new):
            candidate = shuffled_new[new_idx % len(shuffled_new)]
            new_idx += 1
            if new_assign_count[candidate] < new_assign_max_dict[candidate]:
                new_assign_count[candidate] += 1
                replacements[line_idx] = candidate
                found = True
                break
            attempts += 1
            
        if not found:
            candidate = random.choice(shuffled_new)
            new_assign_count[candidate] += 1
            replacements[line_idx] = candidate
            
    for line_idx, new_item in replacements.items():
        old_line = lines[line_idx]
        new_line = pattern.sub(
            lambda m: f"INSERT INTO note_tags (note_id, tag_id) SELECT {m.group(1)}, tag_id FROM tags WHERE tag_name = '{new_item}' AND tag_type = '{tag_type}'",
            old_line
        )
        lines[line_idx] = new_line
        
    return lines

lines = redistribute_type(lines, 'TEACHER', all_teachers, old_teachers, zero_teachers, 3, 5, 1, 2)
lines = redistribute_type(lines, 'DEPARTMENT', all_depts, old_depts, zero_depts, 3, 5, 2, 4)

with open('db/03_init_super_data.sql', 'w', encoding='utf-8') as f:
    f.write('\n'.join(lines))

print("Redistribution complete for TEACHER and DEPARTMENT.")
