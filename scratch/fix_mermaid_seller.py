import os

def process_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    # Replacements
    content = content.replace('developer_id', 'seller_id')
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)

# Process all files
for root, _, files in os.walk('mermaid'):
    for file in files:
        if file.endswith('.txt') or file.endswith('.mermaid'):
            process_file(os.path.join(root, file))

print("Done renaming developer_id to seller_id!")
