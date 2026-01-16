package analytics

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"reviewExplorer/backend/models"
	"sort"
	"strings"
	"time"
	"unicode"
)

type Result struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type ThemePayload struct {
	Count    int      `json:"count"`
	Examples []string `json:"examples"`
}

var keywords map[string][]string

func init() {
	data, err := os.ReadFile("backend/analytics/keywords.json")
	if err != nil {
		data, err = os.ReadFile("analytics/keywords.json")
	}
	if err == nil {
		if err := json.Unmarshal(data, &keywords); err != nil {
			log.Printf("[Analytics] Keywords load error: %v", err)
			return
		}
		normalizeKeywords()
		log.Printf("[Analytics] Keywords loaded: %d groups", len(keywords))
	}
}

func normalizeKeywords() {
	if keywords == nil {
		return
	}
	for cat, kws := range keywords {
		seen := make(map[string]struct{}, len(kws))
		out := make([]string, 0, len(kws))
		for _, kw := range kws {
			kw = normalizeKeyword(kw)
			if kw == "" {
				continue
			}
			if _, ok := seen[kw]; ok {
				continue
			}
			seen[kw] = struct{}{}
			out = append(out, kw)
		}
		keywords[cat] = out
	}
}

func normalizeKeyword(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func splitIntoSentences(text string) []string {
	re := regexp.MustCompile(`[.!?\n]+`)
	parts := re.Split(text, -1)

	var sentences []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) >= 10 {
			sentences = append(sentences, part)
		}
	}
	return sentences
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	return strings.Fields(b.String())
}

type matchInfo struct {
	Score        int
	Specificity  int
	MatchedCount int
}

func matchCategory(tokens []string, category string) matchInfo {
	kws, ok := keywords[category]
	if !ok || len(kws) == 0 || len(tokens) == 0 {
		return matchInfo{}
	}

	matched := make(map[string]struct{}, 4)
	spec := 0
	for _, kw := range kws {
		for _, t := range tokens {

			if strings.HasPrefix(t, kw) {
				if _, ok := matched[kw]; !ok {
					matched[kw] = struct{}{}
					spec += len(kw)
				}
				break
			}
		}
	}

	return matchInfo{
		Score:        len(matched),
		Specificity:  spec,
		MatchedCount: len(matched),
	}
}

func bestCategoryForSentence(sentence string) (string, matchInfo) {
	tokens := tokenize(sentence)
	bestCat := ""
	best := matchInfo{}

	for cat := range keywords {
		mi := matchCategory(tokens, cat)
		if mi.Score == 0 {
			continue
		}
		if mi.Score > best.Score ||
			(mi.Score == best.Score && mi.Specificity > best.Specificity) ||
			(mi.Score == best.Score && mi.Specificity == best.Specificity && (bestCat == "" || cat < bestCat)) {
			best = mi
			bestCat = cat
		}
	}
	return bestCat, best
}

func Analyze(reviews []models.Review) []Result {
	type scoredSentence struct {
		Text        string
		Score       int
		Specificity int
	}

	posCounts := make(map[string]int) 
	negCounts := make(map[string]int)

	posExamples := make(map[string]map[string]scoredSentence) 
	negExamples := make(map[string]map[string]scoredSentence)

	totalPositive, totalNegative, totalNeutral := 0, 0, 0

	type monthStat struct{ Pos, Neg int }
	seasonality := make(map[int]*monthStat) 

	for _, r := range reviews {
		sentences := splitIntoSentences(r.RawText)

		if r.Sentiment == "positive" {
			for _, sentence := range sentences {
				cat, mi := bestCategoryForSentence(sentence)
				if cat == "" {
					continue
				}
				posCounts[cat]++
				if posExamples[cat] == nil {
					posExamples[cat] = make(map[string]scoredSentence)
				}
				existing, ok := posExamples[cat][sentence]
				if !ok || mi.Score > existing.Score || (mi.Score == existing.Score && mi.Specificity > existing.Specificity) {
					posExamples[cat][sentence] = scoredSentence{Text: sentence, Score: mi.Score, Specificity: mi.Specificity}
				}
			}
		} else if r.Sentiment == "negative" {
			for _, sentence := range sentences {
				cat, mi := bestCategoryForSentence(sentence)
				if cat == "" {
					continue
				}
				negCounts[cat]++
				if negExamples[cat] == nil {
					negExamples[cat] = make(map[string]scoredSentence)
				}
				existing, ok := negExamples[cat][sentence]
				if !ok || mi.Score > existing.Score || (mi.Score == existing.Score && mi.Specificity > existing.Specificity) {
					negExamples[cat][sentence] = scoredSentence{Text: sentence, Score: mi.Score, Specificity: mi.Specificity}
				}
			}
		}

		if r.Sentiment == "positive" {
			totalPositive++
		} else if r.Sentiment == "negative" {
			totalNegative++
		} else {
			totalNeutral++
		}

		if len(r.PublishedAt) >= 10 {
			t, err := time.Parse("2006-01-02", r.PublishedAt[:10])
			if err == nil {
				month := int(t.Month())
				if _, ok := seasonality[month]; !ok {
					seasonality[month] = &monthStat{}
				}
				if r.Sentiment == "positive" {
					seasonality[month].Pos++
				} else if r.Sentiment == "negative" {
					seasonality[month].Neg++
				}
			}
		}
	}


	allCategories := make(map[string]bool)
	for cat := range posCounts {
		allCategories[cat] = true
	}
	for cat := range negCounts {
		allCategories[cat] = true
	}

	type balanceItem struct {
		Category string `json:"category"`
		Pos      int    `json:"pos"`
		Neg      int    `json:"neg"`
	}
	var balancePayload []balanceItem
	for cat := range allCategories {
		balancePayload = append(balancePayload, balanceItem{
			Category: cat,
			Pos:      posCounts[cat],
			Neg:      negCounts[cat],
		})
	}
	sort.Slice(balancePayload, func(i, j int) bool {
		return balancePayload[i].Pos+balancePayload[i].Neg > balancePayload[j].Pos+balancePayload[j].Neg
	})

	monthNames := []string{"Янв", "Фев", "Мар", "Апр", "Май", "Июн", "Июл", "Авг", "Сен", "Окт", "Ноя", "Дек"}
	type seasonItem struct {
		Label string `json:"label"`
		Pos   int    `json:"pos"`
		Neg   int    `json:"neg"`
	}
	var seasonPayload []seasonItem
	for i := 1; i <= 12; i++ {
		stat := seasonality[i]
		if stat == nil {
			stat = &monthStat{}
		}
		seasonPayload = append(seasonPayload, seasonItem{
			Label: monthNames[i-1],
			Pos:   stat.Pos,
			Neg:   stat.Neg,
		})
	}


	const maxExamplesPerCategory = 30

	buildThemePayload := func(counts map[string]int, examples map[string]map[string]scoredSentence) map[string]ThemePayload {
		out := make(map[string]ThemePayload, len(counts))
		for cat, cnt := range counts {
			m := examples[cat]
			scored := make([]scoredSentence, 0, len(m))
			for _, s := range m {
				scored = append(scored, s)
			}
			sort.Slice(scored, func(i, j int) bool {
				if scored[i].Score != scored[j].Score {
					return scored[i].Score > scored[j].Score
				}
				if scored[i].Specificity != scored[j].Specificity {
					return scored[i].Specificity > scored[j].Specificity
				}
				return len(scored[i].Text) > len(scored[j].Text)
			})
			limit := maxExamplesPerCategory
			if len(scored) < limit {
				limit = len(scored)
			}
			top := make([]string, 0, limit)
			for i := 0; i < limit; i++ {
				top = append(top, scored[i].Text)
			}
			out[cat] = ThemePayload{Count: cnt, Examples: top}
		}
		return out
	}

	posThemesPayload := buildThemePayload(posCounts, posExamples)
	negThemesPayload := buildThemePayload(negCounts, negExamples)

	return []Result{
		{Name: "Сильные стороны", Type: "bar", Payload: posThemesPayload},
		{Name: "Проблемные зоны", Type: "bar", Payload: negThemesPayload},
		{Name: "Баланс мнений", Type: "stackedBar", Payload: balancePayload},
		{Name: "Сезонность активности", Type: "stackedBar", Payload: seasonPayload},
	}
}
