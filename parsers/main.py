import sys, os, mysql.connector, logging
from datetime import datetime
from scraper_engine import scrape_reviews

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def get_db():
    return mysql.connector.connect(
        user=os.getenv("DB_USER", "root"),
        password=os.getenv("DB_PASSWORD", ""),
        host=os.getenv("DB_HOST", "localhost"),
        database=os.getenv("DB_NAME", "ReviewExplorer"),
        port=int(os.getenv("DB_PORT", 3306))
    )

if __name__ == "__main__":
    if len(sys.argv) < 2: 
        sys.exit(0)
    name = sys.argv[1]
    
    try:
        conn = get_db()
        cursor = conn.cursor()
        cursor.execute("SELECT id FROM schools WHERE full_name LIKE %s OR short_name LIKE %s LIMIT 1", (f"%{name}%", f"%{name}%"))
        row = cursor.fetchone()
        
        if row:
            sid = row[0]
            reviews = scrape_reviews(name)
            if reviews:
                now = datetime.now()
                for r in reviews:
                    cursor.execute("INSERT INTO reviews (school_id, published_at, sentiment, raw_text, parsed_at) VALUES (%s, %s, %s, %s, %s)",
                                 (sid, r["date"], r["rating"], r["text"], now))
                conn.commit()
                logger.info(f" saved {len(reviews)} reviews")
            else:
                logger.info("No reviews found")
        else:
            logger.warning(f"school not in db: {name}")
        conn.close()
    except Exception as e:
        logger.error(f"error {str(e)}")
        sys.exit(1)
