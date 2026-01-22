CREATE TABLE IF NOT EXISTS schools (
    id INT AUTO_INCREMENT PRIMARY KEY,
    org_type VARCHAR(255),
    full_name TEXT,
    short_name TEXT,
    legal_form VARCHAR(255),
    address TEXT,
    website TEXT,
    lat DOUBLE,
    lon DOUBLE
);

CREATE TABLE IF NOT EXISTS reviews (
    id INT AUTO_INCREMENT PRIMARY KEY,
    school_id INT NOT NULL,
    published_at DATE,
    sentiment VARCHAR(20),
    raw_text TEXT,
    parsed_at DATETIME,
    FOREIGN KEY (school_id) REFERENCES schools(id),
    INDEX idx_reviews_school_id (school_id),
    INDEX idx_reviews_sentiment (sentiment)
);

DROP PROCEDURE IF EXISTS sp_delete_reviews;
DROP PROCEDURE IF EXISTS sp_get_review_stats;

DELIMITER //


CREATE PROCEDURE sp_delete_reviews(IN p_school_id INT)
BEGIN
  DELETE FROM reviews WHERE school_id = p_school_id;
END//

CREATE PROCEDURE sp_get_review_stats(IN p_school_id INT)
BEGIN
  SELECT
    COUNT(*) AS total,
    COALESCE(SUM(sentiment = 'positive'), 0) AS positive,
    COALESCE(SUM(sentiment = 'negative'), 0) AS negative
  FROM reviews
  WHERE school_id = p_school_id;
END//


DELIMITER ;
