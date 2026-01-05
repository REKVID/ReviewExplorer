package models

type Review struct {
	ID          int    `json:"id"`
	SchoolID    int    `json:"school_id"`
	PublishedAt string `json:"published_at"`
	Sentiment   string `json:"sentiment"`
	RawText     string `json:"raw_text"`
}
