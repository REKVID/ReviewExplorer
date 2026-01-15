import sys, os, mysql.connector
from datetime import datetime
from scraper_engine import scrape_reviews

if __name__ == "__main__":
    if len(sys.argv) < 2: sys.exit(1)
    
    conn = mysql.connector.connect(
        user=os.getenv("DB_USER", "root"),
        password=os.getenv("DB_PASSWORD", ""),
        host=os.getenv("DB_HOST", "localhost"),
        database=os.getenv("DB_NAME", "ReviewExplorer"),
        port=int(os.getenv("DB_PORT", 3306))
    )
    cursor = conn.cursor()
    
    cursor.execute("SELECT id FROM schools WHERE full_name LIKE %s OR short_name LIKE %s LIMIT 1", (f"%{sys.argv[1]}%", f"%{sys.argv[1]}%"))
    row = cursor.fetchone()
    
    if row:
        for r in scrape_reviews(sys.argv[1]):
            try:
                cursor.execute("INSERT INTO reviews (school_id, published_at, sentiment, raw_text, parsed_at) VALUES (%s, %s, %s, %s, %s)",
                             (row[0], r["date"], r["rating"], r["text"], datetime.now()))
            except: pass
        conn.commit()
    
    conn.close()
