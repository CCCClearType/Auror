import re

with open(r'C:\Users\HP\.gemini\antigravity-ide\brain\4c08fd4d-de5a-47db-b857-8f644e733172\.system_generated\steps\271\content.md', 'r', encoding='utf-8') as f:
    html = f.read()

# Look for <h3 class="font-weight-bold mb-1"> TeacherName
teachers = []
matches = re.finditer(r'<h3 class="font-weight-bold mb-1">\s*([^\s<]+)', html)
for m in matches:
    name = m.group(1).strip()
    if name not in teachers:
        teachers.append(name)

print("Found teachers:", len(teachers))
print(teachers)
