import random
from datetime import datetime, timedelta

# Note IDs available in super_data: 101 to 230
CLASSICS = [110, 120, 130, 140, 150] # 108-1, 25 purchases, old dates
HOT_NEW = [115, 125, 135, 145, 155]  # 114-2, 25 purchases, new dates
OLD_UNPOPULAR = [112, 122, 132, 142, 152] # 109-1, 1 purchase, old dates

def random_date(start_year, end_year):
    start = datetime(start_year, 1, 1)
    end = datetime(end_year, 12, 31)
    delta = end - start
    return start + timedelta(seconds=random.randint(0, int(delta.total_seconds())))

lines = [
    "-- ===========================================================================",
    "-- Patch 05: Eternal Classic Badge Data Generation",
    "-- ===========================================================================\n"
]

all_target_notes = CLASSICS + HOT_NEW + OLD_UNPOPULAR
target_str = ", ".join(map(str, all_target_notes))

# 1. Reset Semester Tags
lines.append(f"DELETE FROM note_tags WHERE note_id IN ({target_str}) AND tag_id IN (SELECT tag_id FROM tags WHERE tag_type = 'SEMESTER');\n")

for n in CLASSICS:
    lines.append(f"INSERT INTO note_tags (note_id, tag_id) SELECT {n}, tag_id FROM tags WHERE tag_name = '108-1' AND tag_type = 'SEMESTER';")
for n in HOT_NEW:
    lines.append(f"INSERT INTO note_tags (note_id, tag_id) SELECT {n}, tag_id FROM tags WHERE tag_name = '114-2' AND tag_type = 'SEMESTER';")
for n in OLD_UNPOPULAR:
    lines.append(f"INSERT INTO note_tags (note_id, tag_id) SELECT {n}, tag_id FROM tags WHERE tag_name = '109-1' AND tag_type = 'SEMESTER';")
lines.append("")

# 2. Clear existing transactions, licenses, reviews for these notes
lines.append(f"DELETE FROM transaction_items WHERE note_id IN ({target_str});")
lines.append(f"DELETE FROM reviews WHERE note_id IN ({target_str});")
lines.append(f"DELETE FROM note_licenses WHERE note_id IN ({target_str});")
lines.append("")

# 3. Generate massive sales
users = list(range(150, 195))
tx_id_counter = 20000
item_id_counter = 20000

def generate_sales(notes, min_sales, max_sales, start_year, end_year):
    global tx_id_counter, item_id_counter
    for note_id in notes:
        num_sales = random.randint(min_sales, max_sales)
        buyers = random.sample(users, num_sales)
        
        for buyer_id in buyers:
            tx_id_counter += 1
            item_id_counter += 1
            buy_date = random_date(start_year, end_year).strftime("%Y-%m-%d %H:%M:%S")
            # 1. Transaction
            lines.append(f"INSERT INTO transactions (transaction_id, user_id, total_amount, transaction_date, receipt_number) VALUES ({tx_id_counter}, {buyer_id}, 500, '{buy_date}', 'REC-CLASSIC-{tx_id_counter}');")
            # 2. Transaction Item
            lines.append(f"INSERT INTO transaction_items (item_id, transaction_id, note_id, purchase_price) VALUES ({item_id_counter}, {tx_id_counter}, {note_id}, 500);")
            # 3. License
            lines.append(f"INSERT INTO note_licenses (user_id, note_id, transaction_item_id, acquired_date, status) VALUES ({buyer_id}, {note_id}, {item_id_counter}, '{buy_date}', 'ACTIVE');")
            # 4. Review
            if random.random() < 0.8:
                attitude = random.choices(['POSITIVE', 'NEGATIVE'], weights=[0.9, 0.1])[0]
                review_date = (datetime.strptime(buy_date, "%Y-%m-%d %H:%M:%S") + timedelta(days=random.randint(1, 10))).strftime("%Y-%m-%d %H:%M:%S")
                lines.append(f"INSERT INTO reviews (note_id, user_id, attitude, content, created_at, status) VALUES ({note_id}, {buyer_id}, '{attitude}', '這份筆記真的很棒！對期末考幫助極大，強力推薦！', '{review_date}', 'VISIBLE');")

lines.append("-- Inserting sales for CLASSICS")
generate_sales(CLASSICS, 20, 30, 2021, 2024)
lines.append("\n-- Inserting sales for HOT NEW")
generate_sales(HOT_NEW, 20, 30, 2026, 2026)
lines.append("\n-- Inserting sales for OLD UNPOPULAR")
generate_sales(OLD_UNPOPULAR, 1, 2, 2021, 2022)

lines.append(f"\nSELECT setval('transactions_transaction_id_seq', (SELECT MAX(transaction_id) FROM transactions));")
lines.append(f"SELECT setval('transaction_items_item_id_seq', (SELECT MAX(item_id) FROM transaction_items));")

with open('db/05_patch_classics.sql', 'w', encoding='utf-8') as f:
    f.write("\n".join(lines) + "\n")

print("Generated db/05_patch_classics.sql")
