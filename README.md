# ReviewExplorer

A web application for monitoring and analyzing educational institutions based on open data and automated review collection. The system performs sentiment analysis, thematic categorization, and temporal pattern detection on school reviews.

## Architecture Overview

ReviewExplorer follows a microservices-oriented architecture with clear separation of concerns:

```
┌─────────────┐
│   Frontend  │  Vanilla JavaScript, Chart.js, 2GIS Maps API
└──────┬──────┘
       │ HTTP/REST
┌──────▼─────────────────────────────────────┐
│         Go Backend (Chi Router)            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │ Handlers │  │ Analytics│  │   MySQL  │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  │
└───────┼─────────────┼──────────────┼───────┘
        │            │              │
        │            │              │
┌───────▼────────────▼──────────────▼───────┐
│  Python Scraper (Selenium + Chrome)       │
│  Executed via exec.Command on demand      │
└───────────────────────────────────────────┘
```

### Core Components

**Backend (Go 1.23)**
- HTTP server using Chi router with middleware stack (logger, recovery, CORS)
- Database layer with connection pooling via `database/sql`
- Analytics engine implementing keyword-based thematic analysis
- Request handlers for schools, analysis, and data refresh operations

**Database (MySQL)**
- Normalized schema with foreign key constraints
- Stored procedures for review statistics and deletion operations
- Indexed queries on `school_id` and `sentiment` columns
- Support for geospatial data (lat/lon coordinates)

**Scraper (Python 3 + Selenium)**
- Headless Chrome automation for web scraping
- Anti-detection measures (user-agent spoofing, webdriver property masking)
- Review extraction with deduplication logic
- Direct MySQL insertion via `mysql-connector-python`

**Frontend (Vanilla JavaScript)**
- Client-side routing and state management
- Chart.js integration for data visualization
- 2GIS Maps API for geospatial rendering
- Particles.js for visual effects
- Responsive layout with CSS Grid

## Technical Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Backend | Go 1.23, Chi Router | HTTP API, business logic |
| Database | MySQL 8.0+ | Data persistence, stored procedures |
| Scraper | Python 3, Selenium, Chrome | Web scraping automation |
| Frontend | Vanilla JS, Chart.js, 2GIS | UI, visualization, mapping |
| Deployment | Docker, Docker Compose | Containerization, orchestration |

## Data Flow

### 1. School Data Loading
CSV files from Mos.ru open data portal are processed by `backend/cmd/loader.go`:
- Parses semicolon-delimited CSV with 26+ columns
- Extracts geospatial coordinates from GeoJSON-formatted strings
- Bulk inserts into `schools` table with `INSERT IGNORE` for deduplication

### 2. Review Collection
On-demand scraping triggered via `/analyze` endpoint:
- Backend executes Python scraper as subprocess
- Scraper performs Yandex search with `site:schoolotzyv.ru` query
- Extracts review URLs, navigates to target pages
- Parses review text, dates, and sentiment labels
- Inserts into `reviews` table with school_id foreign key

### 3. Analytics Processing
The analytics engine (`backend/analytics/engine.go`) implements:

**Keyword Matching Algorithm**
- Tokenization with Unicode-aware normalization
- Prefix matching against categorized keyword sets
- Scoring based on match count and keyword specificity
- Sentence-level categorization with best-match selection

**Thematic Analysis**
- Positive/negative theme extraction from sentence-level sentiment
- Example selection using score-based ranking (top 30 per category)
- Category balance calculation (positive vs negative mentions)

**Temporal Analysis**
- Month-based seasonality detection from `published_at` timestamps
- Aggregation of positive/negative counts per month
- Time-series data for stacked bar visualization

**Output Structure**
```json
{
  "school_name": "string",
  "stats": {
    "total": 0,
    "positive": 0,
    "negative": 0,
    "neutral": 0
  },
  "analytics": [
    {
      "name": "Сильные стороны",
      "type": "bar",
      "payload": { "category": { "count": 0, "examples": [] } }
    },
    {
      "name": "Проблемные зоны",
      "type": "bar",
      "payload": { "category": { "count": 0, "examples": [] } }
    },
    {
      "name": "Баланс мнений",
      "type": "stackedBar",
      "payload": [{ "category": "string", "pos": 0, "neg": 0 }]
    },
    {
      "name": "Сезонность активности",
      "type": "stackedBar",
      "payload": [{ "label": "Янв", "pos": 0, "neg": 0 }]
    }
  ]
}
```

## Database Schema

```sql
schools (
  id INT PRIMARY KEY AUTO_INCREMENT,
  org_type VARCHAR(255),
  full_name TEXT,
  short_name TEXT,
  legal_form VARCHAR(255),
  address TEXT,
  website TEXT,
  lat DOUBLE,
  lon DOUBLE
)

reviews (
  id INT PRIMARY KEY AUTO_INCREMENT,
  school_id INT NOT NULL,
  published_at DATE,
  sentiment VARCHAR(20),
  raw_text TEXT,
  parsed_at DATETIME,
  FOREIGN KEY (school_id) REFERENCES schools(id),
  INDEX idx_reviews_school_id (school_id),
  INDEX idx_reviews_sentiment (sentiment)
)
```

**Stored Procedures**
- `sp_get_review_stats(school_id)`: Aggregates total, positive, negative counts
- `sp_delete_reviews(school_id)`: Removes all reviews for a school (used in refresh)

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/schools?q=<query>` | Search schools by name (LIKE pattern matching) |
| POST | `/analyze` | Trigger analysis: scrape if needed, return analytics |
| POST | `/refresh` | Delete existing reviews and re-scrape |
| GET | `/*` | Static file serving (frontend) |

**Request/Response Examples**

```bash
# Search schools
GET /schools?q=гимназия
Response: [{ "id": 1, "full_name": "...", "lat": 55.75, "lon": 37.61, ... }]

# Analyze school
POST /analyze
Body: { "query": "ГБОУ Школа №123" }
Response: { "school_name": "...", "stats": {...}, "analytics": [...] }
```

## Docker Deployment

**Dockerfile Strategy**
- Multi-stage approach: base image includes Go 1.23, Python 3, Chromium
- System dependencies installed via `apt-get` (chromium, chromium-driver, libnss3)
- Python packages: `selenium`, `mysql-connector-python`
- Go binary compiled at build time
- Single container runs both Go server and Python scraper

**Docker Compose Configuration**
```yaml
services:
  app:
    build: .
    ports:
      - "8081:8081"
    env_file: .env
    volumes:
      - ./frontend:/app/frontend  # Hot-reload for development
    extra_hosts:
      - "host.docker.internal:host-gateway"  # Database access
```

**Environment Variables**
```
DB_USER=root
DB_PASSWORD=password
DB_HOST=host.docker.internal
DB_NAME=ReviewExplorer
```

## Design Decisions

### Why Go for Backend?
- Native concurrency with goroutines for handling multiple requests
- Low memory footprint suitable for containerized environments
- Fast compilation and binary distribution
- Strong standard library for HTTP and database operations
- Type safety reduces runtime errors

### Why Python for Scraping?
- Selenium ecosystem is mature and well-documented
- Rapid development for parsing and text extraction
- Easy integration with ChromeDriver
- Subprocess execution from Go allows isolation and error handling

### Why MySQL?
- ACID compliance for data integrity
- Stored procedures reduce application-level query complexity
- Foreign key constraints ensure referential integrity
- Indexed queries provide fast lookups on large datasets

### Why Vanilla JavaScript?
- No build step or transpilation overhead
- Direct DOM manipulation for performance
- Minimal dependencies (Chart.js, 2GIS, Particles.js via CDN)
- Easy debugging and deployment

### Analytics Engine Design
- **Keyword-based approach** instead of ML: interpretable results, no training data required, fast execution
- **Sentence-level analysis**: granular categorization allows multiple themes per review
- **Scoring system**: specificity-weighted matching prevents false positives from short keywords
- **Example selection**: top-scored sentences provide representative samples for each category

## Performance Considerations

- Database connection pooling via `sql.DB` with default settings
- Indexed queries on `school_id` and `sentiment` columns
- Lazy scraping: reviews collected only when requested
- Cached analytics: results computed on-demand, not pre-aggregated
- Frontend debouncing: search suggestions delayed by 300ms to reduce API calls

## Security Measures

- CORS middleware restricts cross-origin requests (currently permissive for development)
- SQL injection prevention via parameterized queries (`?` placeholders)
- Environment variable isolation for database credentials
- Input validation on query parameters (minimum length checks)

## Development Workflow

1. **Database Setup**: Run `migrations.sql` to create schema and procedures
2. **Data Loading**: Execute `go run backend/cmd/loader.go` to import CSV data
3. **Backend Development**: `go run backend/main.go` (requires `.env` file)
4. **Frontend Development**: Serve `frontend/` directory via any HTTP server
5. **Docker Build**: `docker-compose up --build` for containerized deployment

## File Structure

```
ReviewExplorer/
├── backend/
│   ├── main.go              # HTTP server entry point
│   ├── handlers/            # Request handlers
│   │   ├── schools.go       # School search endpoint
│   │   └── analytics.go    # Analysis and refresh endpoints
│   ├── analytics/
│   │   ├── engine.go       # Thematic analysis engine
│   │   └── keywords.json   # Category keyword definitions
│   ├── db/
│   │   └── mysql.go        # Database connection
│   ├── models/
│   │   ├── school.go       # School data model
│   │   └── review.go      # Review data model
│   ├── cmd/
│   │   └── loader.go       # CSV data importer
│   └── migrations.sql      # Database schema
├── frontend/
│   ├── index.html          # Main application page
│   ├── app.js              # Client-side logic
│   ├── style.css           # Styling
│   └── about.html          # Project information
├── parsers/
│   ├── main.py             # Scraper entry point
│   └── scraper_engine.py   # Selenium scraping logic
├── data/
│   └── *.csv               # Open data files
├── Dockerfile              # Container build instructions
├── docker-compose.yml      # Orchestration configuration
└── go.mod                  # Go module dependencies
```

## Future Enhancements

- Redis caching for analytics results
- Background job queue for asynchronous scraping
- Machine learning-based sentiment analysis
- Real-time WebSocket updates for long-running operations
- Multi-language support for international datasets
- GraphQL API for flexible data querying

## License

This project uses open data from [Mos.ru Open Data Portal](https://data.mos.ru/opendata/2263) and scrapes publicly available reviews from schoolotzyv.ru.
