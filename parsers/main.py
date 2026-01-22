import sys, os, mysql.connector
from datetime import datetime
from scraper_engine import scrape_reviews

if __name__ == "__main__":
    if len(sys.argv) < 2: sys.exit(1)
    school_name = sys.argv[1]
    
    conn = mysql.connector.connect(
        user=os.getenv("DB_USER", "root"),
        password=os.getenv("DB_PASSWORD", ""),
        host=os.getenv("DB_HOST", "localhost"),
        database=os.getenv("DB_NAME", "ReviewExplorer"),
        port=int(os.getenv("DB_PORT", 3306))
    )
    cursor = conn.cursor()
    
    print(f"[parser] start: {school_name}")
    cursor.execute(
        "SELECT id FROM schools WHERE full_name LIKE %s OR short_name LIKE %s LIMIT 1",
        (f"%{school_name}%", f"%{school_name}%")
    )
    row = cursor.fetchone()
    
    if row:
        inserted = 0
        scraped = 0
        for r in scrape_reviews(school_name):
            scraped += 1
            try:
                cursor.execute("INSERT INTO reviews (school_id, published_at, sentiment, raw_text, parsed_at) VALUES (%s, %s, %s, %s, %s)",
                             (row[0], r["date"], r["rating"], r["text"], datetime.now()))
                inserted += 1
            except Exception as e:
                print(f"[parser] insert error: {e}")
        conn.commit()
        print(f"[parser] done: scraped={scraped} inserted={inserted} school_id={row[0]}")
    else:
        print("[parser] school not found in DB")
    
    conn.close()
