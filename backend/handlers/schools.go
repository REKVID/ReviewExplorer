package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reviewExplorer/backend/analytics"
	"reviewExplorer/backend/db"
	"reviewExplorer/backend/models"
)

type AnalyzeRequest struct {
	Query string `json:"query"`
}

func GetSchools(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		json.NewEncoder(w).Encode([]string{})
		return
	}

	rows, err := db.DB.Query("SELECT id, full_name, short_name FROM schools WHERE full_name LIKE ? OR short_name LIKE ? LIMIT 5",
		"%"+query+"%", "%"+query+"%")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var schools []models.School
	for rows.Next() {
		var s models.School
		if err := rows.Scan(&s.ID, &s.FullName, &s.ShortName); err != nil {
			continue
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

	var id int
	var fullName string
	err := db.DB.QueryRow("SELECT id, full_name FROM schools WHERE full_name LIKE ? OR short_name LIKE ? LIMIT 1",
		"%"+req.Query+"%", "%"+req.Query+"%").Scan(&id, &fullName)

	if err != nil {
		log.Printf("not found : %s", req.Query)
		http.Error(w, "Школа не найдена", http.StatusNotFound)
		return
	}

	var reviewCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM reviews WHERE school_id = ?", id).Scan(&reviewCount)

	if reviewCount == 0 {
		log.Printf("No reviews - starting parser")
		cmd := exec.Command("python3", "parsers/main.py", fullName)
		cmd.Env = os.Environ()
		output, err := cmd.CombinedOutput()
		log.Printf("Parser result:\n%s", string(output))

		if err != nil {
			http.Error(w, "Ошибка при сборе отзывов: "+err.Error(), http.StatusInternalServerError)
			return
		}
		db.DB.QueryRow("SELECT COUNT(*) FROM reviews WHERE school_id = ?", id).Scan(&reviewCount)
	}

	var reviews []models.Review
	var pos, neg int
	rows, _ := db.DB.Query("SELECT id, school_id, raw_text, sentiment FROM reviews WHERE school_id = ?", id)
	for rows.Next() {
		var rev models.Review
		rows.Scan(&rev.ID, &rev.SchoolID, &rev.RawText, &rev.Sentiment)
		reviews = append(reviews, rev)
		if rev.Sentiment == "positive" {
			pos++
		}
		if rev.Sentiment == "negative" {
			neg++
		}
	}
	defer rows.Close()

	response := map[string]interface{}{
		"school_name": fullName,
		"stats": map[string]int{
			"total":    len(reviews),
			"positive": pos,
			"negative": neg,
			"neutral":  len(reviews) - pos - neg,
		},
		"analytics": analytics.Analyze(reviews),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
