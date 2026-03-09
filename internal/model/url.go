package model

import "time"

type URL struct {
	Code        string    `json:"code"`
	Original    string    `json:"original"`
	CreatedAt   time.Time `json:"created_at"`
	IP          string    `json:"ip"`
	DeleteToken string    `json:"delete_token"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Short          string `json:"short"`
	Code           string `json:"code"`
	DeleteToken    string `json:"delete_token"`
	ActiveLinks    int    `json:"active_links"`
	DailyCreations int    `json:"daily_creations"`
	MaxLinks       int    `json:"max_links"`
	MaxDaily       int    `json:"max_daily"`
}
