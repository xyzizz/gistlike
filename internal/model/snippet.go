package model

import "time"

var AllowedLanguages = []string{
	"plaintext",
	"go",
	"python",
	"javascript",
	"sql",
	"json",
}

type Snippet struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Language    string    `json:"language"`
	Content     string    `json:"content"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SnippetInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Content     string `json:"content"`
	IsPublic    bool   `json:"is_public"`
}

type SnippetFilter struct {
	Query string
	Limit int
}
