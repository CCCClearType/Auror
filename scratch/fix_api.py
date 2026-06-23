import re
with open('api/api_list.txt', 'r', encoding='utf-8') as f:
    content = f.read()
content = content.replace('/api/games', '/api/notes')
content = content.replace('/developer/games', '/seller/notes')
content = content.replace('/admin/games', '/admin/notes')
content = content.replace('{game_id}', '{note_id}')
content = content.replace('/developer/tags', '/seller/tags')
content = content.replace('Games', 'Notes')
with open('api/api_list.txt', 'w', encoding='utf-8') as f:
    f.write(content)
