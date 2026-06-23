import os
import glob
import re

def process_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    # Replacements
    content = content.replace('games', 'notes')
    content = content.replace('game_id', 'note_id')
    content = content.replace('game_tags', 'note_tags')
    content = content.replace('game_media', 'note_media')
    content = content.replace('game_licenses', 'note_licenses')
    content = content.replace('game', 'note')
    content = content.replace('Games', 'Notes')
    content = content.replace('Game', 'Note')
    content = content.replace('GAMES', 'NOTES')
    content = content.replace('GAME', 'NOTE')
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)

# Rename files
base_dir = 'mermaid/mermaid_modules'
renames = {
    '01_User_Game_Permission.mermaid': '01_User_Note_Permission.mermaid',
    '02_Game_Tag_Media.mermaid': '02_Note_Tag_Media.mermaid',
    '05_Game_License.mermaid': '05_Note_License.mermaid'
}

for old, new in renames.items():
    old_path = os.path.join(base_dir, old)
    new_path = os.path.join(base_dir, new)
    if os.path.exists(old_path):
        os.rename(old_path, new_path)

# Process all files
for root, _, files in os.walk('mermaid'):
    for file in files:
        if file.endswith('.txt') or file.endswith('.mermaid'):
            process_file(os.path.join(root, file))

print("Done renaming!")
