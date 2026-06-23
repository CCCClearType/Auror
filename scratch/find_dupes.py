import os

for root, dirs, files in os.walk('frontend'):
    for f in files:
        if f.endswith('.html'):
            filepath = os.path.join(root, f)
            try:
                content = open(filepath, 'r', encoding='utf-8').read()
                count = content.count('src="/assets/js/main.js"')
                if count > 1:
                    print(f"Duplicate main.js in: {filepath} ({count} times)")
            except Exception as e:
                pass
