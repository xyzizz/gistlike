package handler

import "gistlike/internal/model"

type ViewData struct {
	AppName      string
	PageTitle    string
	PageTemplate string
	Scripts      []string

	SearchQuery string
	Snippets    []model.Snippet
	Snippet     *model.Snippet
	Languages   []string
	FormMode    string

	ErrorCode    int
	ErrorTitle   string
	ErrorMessage string
}
