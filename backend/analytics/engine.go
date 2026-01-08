package analytics

import (
	"encoding/json"
	"log"
	"os"
	"reviewExplorer/backend/models"
	"strings"
)

type Result struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

var keywords map[string][]string

func init() {
	data, err := os.ReadFile("backend/analytics/keywords.json")
	if err != nil {
		data, err = os.ReadFile("analytics/keywords.json")
	}
	if err == nil {
		json.Unmarshal(data, &keywords)
		log.Printf("[Analytics] Keywords loaded: %d groups", len(keywords))
	} else {
		log.Printf("[Analytics] Failed to load keywords: %v", err)
	}
}

func GetTheme(text string) string {
	text = strings.ToLower(text)
	for theme, kws := range keywords {
		for _, kw := range kws {
			if strings.Contains(text, kw) {
				return theme
			}
		}
	}
	return "Другое"
}

func Analyze(reviews []models.Review) []Result {
	themes := make(map[string]int)
	depth := map[string]float64{"Позитивные": 0, "Отрицательные": 0}
	posCount, negCount := 0, 0

	for _, r := range reviews {
		words := float64(len(strings.Fields(r.RawText)))
		if r.Sentiment == "positive" {
			depth["Позитивные"] += words
			posCount++
			themes[GetTheme(r.RawText)]++
		} else if r.Sentiment == "negative" {
			depth["Отрицательные"] += words
			negCount++
		}
	}

	if posCount > 0 {
		depth["Позитивные"] /= float64(posCount)
	}
	if negCount > 0 {
		depth["Отрицательные"] /= float64(negCount)
	}

	return []Result{
		{Name: "Темы отзывов", Type: "bar", Payload: themes},
		{Name: "Средняя длина (слов)", Type: "bar", Payload: depth},
	}
}
