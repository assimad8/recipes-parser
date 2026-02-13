package entities

import "time"

type Recipe struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	ImageURL  string    `json:"image_url"`
	Link      string    `json:"link"`
}