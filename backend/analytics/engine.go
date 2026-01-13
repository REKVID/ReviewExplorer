package analytics

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reviewExplorer/backend/models"
	"sort"
	"strings"
	"time"
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
	return "Общие впечатления"
}

func Analyze(reviews []models.Review) []Result {
	posThemes := make(map[string][]string)
	negThemes := make(map[string][]string)
	depth := map[string]float64{"Положительные": 0, "Отрицательные": 0}
	posCount, negCount := 0, 0

	type stat struct{ Pos, Neg int }
	dynamics := make(map[string]*stat)

	for _, r := range reviews {
		words := float64(len(strings.Fields(r.RawText)))
		theme := GetTheme(r.RawText)

		if r.Sentiment == "positive" {
			depth["Положительные"] += words
			posCount++
			posThemes[theme] = append(posThemes[theme], r.RawText)
		} else if r.Sentiment == "negative" {
			depth["Отрицательные"] += words
			negCount++
			negThemes[theme] = append(negThemes[theme], r.RawText)
		}

		if len(r.PublishedAt) >= 10 {
			t, _ := time.Parse("2006-01-02", r.PublishedAt[:10])
			key := fmt.Sprintf("%d", t.Year())
			if _, ok := dynamics[key]; !ok {
				dynamics[key] = &stat{}
			}
			if r.Sentiment == "positive" {
				dynamics[key].Pos++
			} else if r.Sentiment == "negative" {
				dynamics[key].Neg++
			}
		}
	}

	if posCount > 0 {
		depth["Положительные"] /= float64(posCount)
	}
	if negCount > 0 {
		depth["Отрицательные"] /= float64(negCount)
	}

	var dynKeys []string
	for k := range dynamics {
		dynKeys = append(dynKeys, k)
	}
	sort.Strings(dynKeys)

	type pt struct {
		Label string  `json:"label"`
		Value float64 `json:"value"`
	}
	var dynPayload []pt
	for _, k := range dynKeys {
		total := dynamics[k].Pos + dynamics[k].Neg
		val := 0.0
		if total > 0 {
			val = float64(dynamics[k].Pos) / float64(total) * 100
		}
		dynPayload = append(dynPayload, pt{k, val})
	}

	return []Result{
		{Name: "Сильные стороны", Type: "bar", Payload: posThemes},
		{Name: "Проблемные зоны", Type: "bar", Payload: negThemes},
		{Name: "Средняя длина отзыва (слов)", Type: "bar", Payload: depth},
		{Name: "Динамика удовлетворенности (%)", Type: "line", Payload: dynPayload},
	}
}
