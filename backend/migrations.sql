CREATE TABLE IF NOT EXISTS schools (
    id INT AUTO_INCREMENT PRIMARY KEY,
    org_type VARCHAR(255),
    full_name TEXT,
    short_name TEXT,
    legal_form VARCHAR(255),
    address TEXT,
    website TEXT
);

CREATE TABLE IF NOT EXISTS reviews (
    id INT AUTO_INCREMENT PRIMARY KEY,
    school_id INT NOT NULL,
    published_at DATE,
    sentiment VARCHAR(20),
    raw_text TEXT,
    parsed_at DATETIME,
    FOREIGN KEY (school_id) REFERENCES schools(id)
);
