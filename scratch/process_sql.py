import re

with open(r'c:\Users\HP\Downloads\dbms-git\Ilearn\db\03_init_super_data.sql', 'r', encoding='utf-8') as f:
    lines = f.readlines()

new_lines = []
note_tags = []

for line in lines:
    if 'INSERT INTO notes' in line:
        # Extract note_id and semester
        # Pattern: INSERT INTO notes (note_id, seller_id, title, description, price, status, semester) VALUES (250, 153, '...', '...', 1253.15, 'ACTIVE', '113-1');
        match = re.search(r'INSERT INTO notes \((.*?)\) VALUES \((.*?)\);', line)
        if match:
            cols = match.group(1).split(', ')
            vals = match.group(2).split(', ')
            
            # Find note_id index
            note_id_idx = cols.index('note_id')
            note_id = vals[note_id_idx]
            
            # Find semester index
            if 'semester' in cols:
                sem_idx = cols.index('semester')
                sem_val = vals[sem_idx].strip("'")
                
                # Assume 113-1 is tag_id = 5
                tag_id = 5
                if sem_val == '112-2': tag_id = 6
                if sem_val == '112-1': tag_id = 7
                
                note_tags.append((note_id, tag_id))
                
                # Remove semester from cols and vals
                cols.pop(sem_idx)
                vals.pop(sem_idx)
            
            new_line = f"INSERT INTO notes ({', '.join(cols)}) VALUES ({', '.join(vals)});\n"
            new_lines.append(new_line)
        else:
            new_lines.append(line)
    else:
        new_lines.append(line)

# Fetch teachers
with open(r'C:\Users\HP\.gemini\antigravity-ide\brain\4c08fd4d-de5a-47db-b857-8f644e733172\.system_generated\steps\271\content.md', 'r', encoding='utf-8') as f:
    html = f.read()

teachers = []
matches = re.finditer(r'<h3 class="font-weight-bold mb-1">\s*([^\s<]+)', html)
for m in matches:
    name = m.group(1).strip()
    if name not in teachers:
        teachers.append(name)

# Append teacher tags and note_tags to the end
new_lines.append("\n-- Insert Teacher Tags\n")
new_lines.append("INSERT INTO tags (tag_name, tag_type) VALUES\n")
teacher_vals = [f"('{t}', 'TEACHER')" for t in teachers]
new_lines.append(",\n".join(teacher_vals) + ";\n")

new_lines.append("\n-- Insert Semester Note Tags for Super Data\n")
new_lines.append("INSERT INTO note_tags (note_id, tag_id) VALUES\n")
nt_vals = [f"({n}, {t})" for n, t in note_tags]
# split into chunks of 1000
chunk_size = 1000
for i in range(0, len(nt_vals), chunk_size):
    chunk = nt_vals[i:i+chunk_size]
    if i > 0:
        new_lines.append("INSERT INTO note_tags (note_id, tag_id) VALUES\n")
    new_lines.append(",\n".join(chunk) + ";\n")


with open(r'c:\Users\HP\Downloads\dbms-git\Ilearn\db\03_init_super_data.sql', 'w', encoding='utf-8') as f:
    f.writelines(new_lines)

print("Done processing 03_init_super_data.sql")
