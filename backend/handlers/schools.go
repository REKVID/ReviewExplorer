package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reviewExplorer/backend/analytics"
	"reviewExplorer/backend/db"
	"reviewExplorer/backend/models"
	"strings"
)

type AnalyzeRequest struct {
	Query string `json:"query"`
}

type RefreshRequest struct {
	Query string `json:"query"`
}

func getReviewStats(schoolID int) (total, pos, neg int, err error) {
	err = db.DB.QueryRow("CALL sp_get_review_stats(?)", schoolID).Scan(&total, &pos, &neg)
	if err == nil {
		return total, pos, neg, nil
	}

	err = db.DB.QueryRow(`
		SELECT
			COUNT(*) AS total,
			COALESCE(SUM(sentiment = 'positive'), 0) AS positive,
			COALESCE(SUM(sentiment = 'negative'), 0) AS negative
		FROM reviews
		WHERE school_id = ?
	`, schoolID).Scan(&total, &pos, &neg)
	return total, pos, neg, err
}

func findSchoolByQuery(q string) (id int, fullName string, shortName string, err error) {
	err = db.DB.QueryRow(
		"SELECT id, full_name, short_name FROM schools WHERE full_name LIKE ? OR short_name LIKE ? LIMIT 1",
		"%"+q+"%", "%"+q+"%",
	).Scan(&id, &fullName, &shortName)
	return
}

func GetSchools(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	var rows *sql.Rows
	var err error

	if query == "" {
		rows, err = db.DB.Query(`
			SELECT s.id, s.full_name, s.short_name, s.lat, s.lon,
			       EXISTS(SELECT 1 FROM reviews r WHERE r.school_id = s.id) AS review_count
			FROM schools s
			WHERE s.lat IS NOT NULL
			LIMIT 200
		`)
	} else {
		rows, err = db.DB.Query(`
			SELECT s.id, s.full_name, s.short_name, s.lat, s.lon,
			       EXISTS(SELECT 1 FROM reviews r WHERE r.school_id = s.id) AS review_count
			FROM schools s
			WHERE s.full_name LIKE ? OR s.short_name LIKE ?
			LIMIT 5
		`, "%"+query+"%", "%"+query+"%")
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var schools []models.School
	for rows.Next() {
		var s models.School
		var lat, lon sql.NullFloat64
		if err := rows.Scan(&s.ID, &s.FullName, &s.ShortName, &lat, &lon, &s.ReviewCnt); err != nil {
			continue
		}
		if lat.Valid {
			s.Lat = lat.Float64
			s.Lon = lon.Float64
		}
		schools = append(schools, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schools)
}

func Analyze(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, fullName, shortName, err := findSchoolByQuery(req.Query)
	if err != nil {
		log.Printf("not found : %s", req.Query)
		http.Error(w, "Школа не найдена", http.StatusNotFound)
		return
	}

	total, pos, neg, err := getReviewStats(id)
	if err != nil {
		http.Error(w, "Ошибка получения статистики из БД", http.StatusInternalServerError)
		return
	}

	if total == 0 {
		parserQuery := strings.TrimSpace(shortName)
		if parserQuery == "" {
			parserQuery = fullName
		}
		log.Printf("parser start: %s (school_id=%d)", fullName, id)
		cmd := exec.Command("python3", "parsers/main.py", parserQuery)
		cmd.Env = os.Environ()
		output, err := cmd.CombinedOutput()
		log.Printf("parser output:\n%s", string(output))

		if err != nil {
			log.Printf("parser error: %v", err)
			http.Error(w, "Ошибка при сборе отзывов: "+err.Error(), http.StatusInternalServerError)
			return
		}
		total, pos, neg, err = getReviewStats(id)
		if err != nil {
			http.Error(w, "Ошибка получения статистики из БД", http.StatusInternalServerError)
			return
		}
	}

	var reviews []models.Review
	rows, qErr := db.DB.Query("SELECT id, school_id, published_at, raw_text, sentiment FROM reviews WHERE school_id = ?", id)
	if qErr != nil {
		http.Error(w, "Ошибка чтения отзывов из БД", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var rev models.Review
		if scanErr := rows.Scan(&rev.ID, &rev.SchoolID, &rev.PublishedAt, &rev.RawText, &rev.Sentiment); scanErr != nil {
			continue
		}
		reviews = append(reviews, rev)
	}

	response := map[string]interface{}{
		"school_name": fullName,
		"stats": map[string]int{
			"total":    total,
			"positive": pos,
			"negative": neg,
			"neutral":  total - pos - neg,
		},
		"analytics": analytics.Analyze(reviews),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, fullName, shortName, err := findSchoolByQuery(req.Query)
	if err != nil {
		http.Error(w, "Школа не найдена", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if _, delErr := db.DB.Exec("CALL sp_delete_reviews(?)", id); delErr != nil {
		if _, fallbackDelErr := db.DB.Exec("DELETE FROM reviews WHERE school_id = ?", id); fallbackDelErr != nil {
			http.Error(w, "Ошибка очистки отзывов", http.StatusInternalServerError)
			return
		}
	}
	parserQuery := strings.TrimSpace(shortName)
	if parserQuery == "" {
		parserQuery = fullName
	}
	cmd := exec.Command("python3", "parsers/main.py", fullName)
	cmd = exec.Command("python3", "parsers/main.py", parserQuery)
	cmd.Env = os.Environ()
	if out, runErr := cmd.CombinedOutput(); runErr != nil {
		log.Printf("parser output:\n%s", string(out))
		log.Printf("parser error: %v", runErr)
		http.Error(w, "Ошибка при обновлении отзывов: "+runErr.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
