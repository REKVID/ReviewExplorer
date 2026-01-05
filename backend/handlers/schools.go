package handlers

import (
	"encoding/json"
	"net/http"
	"reviewExplorer/backend/db"
	"reviewExplorer/backend/models"
)

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
