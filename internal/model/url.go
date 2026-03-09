package model

import "time"

type URL struct {
	Code      string    `json:"code"`
	Original  string    `json:"original"`
	CreatedAt time.Time `json:"created_at"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Short string `json:"short"`
	Code  string `json:"code"`
}
