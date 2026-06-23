import re

with open(r'c:\Users\HP\Downloads\dbms-git\Ilearn\db\03_init_super_data.sql', 'r', encoding='utf-8') as f:
    content = f.read()

content = re.sub(r'INSERT INTO notes \(note_id, seller_id, title, description, price, status\) VALUES \((.*?)\);', r"INSERT INTO notes (note_id, seller_id, title, description, price, status, semester) VALUES (\1, '113-1');", content)

with open(r'c:\Users\HP\Downloads\dbms-git\Ilearn\db\03_init_super_data.sql', 'w', encoding='utf-8') as f:
    f.write(content)
