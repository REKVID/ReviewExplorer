package handlers

import (
	"encoding/json"
	"net/http"
	"reviewExplorer/backend/db"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func GetAnalytics(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	var total, positive, negative int

	db.DB.QueryRow("SELECT COUNT(*) FROM reviews WHERE school_id = ?", id).Scan(&total)
	db.DB.QueryRow("SELECT COUNT(*) FROM reviews WHERE school_id = ? AND sentiment = 'positive'", id).Scan(&positive)
	db.DB.QueryRow("SELECT COUNT(*) FROM reviews WHERE school_id = ? AND sentiment = 'negative'", id).Scan(&negative)

	response := map[string]interface{}{
		"stats": map[string]int{
			"positive": positive,
			"negative": negative,
			"neutral":  total - positive - negative,
			"total":    total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
