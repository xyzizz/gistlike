package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"gistlike/internal/model"
	"gistlike/internal/repository"
)

type ValidationError struct {
	Fields map[string]string `json:"fields"`
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

type SnippetService struct {
	repo repository.SnippetRepository
	now  func() time.Time
}

func NewSnippetService(repo repository.SnippetRepository) *SnippetService {
	return &SnippetService{
		repo: repo,
		now:  time.Now,
	}
}

func (s *SnippetService) AllowedLanguages() []string {
	languages := make([]string, len(model.AllowedLanguages))
	copy(languages, model.AllowedLanguages)
	return languages
}

func (s *SnippetService) Create(ctx context.Context, input model.SnippetInput) (*model.Snippet, error) {
	normalized, err := s.validateInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now().UTC()
	snippet := &model.Snippet{
		ID:          uuid.NewString(),
		Title:       normalized.Title,
		Description: normalized.Description,
		Language:    normalized.Language,
		Content:     normalized.Content,
		IsPublic:    normalized.IsPublic,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, snippet); err != nil {
		return nil, err
	}

	return snippet, nil
}

func (s *SnippetService) GetByID(ctx context.Context, id string) (*model.Snippet, error) {
	return s.repo.GetByID(ctx, strings.TrimSpace(id))
}

func (s *SnippetService) ListPublic(ctx context.Context, query string) ([]model.Snippet, error) {
	return s.repo.ListPublic(ctx, model.SnippetFilter{
		Query: strings.TrimSpace(query),
		Limit: 100,
	})
}

func (s *SnippetService) Update(ctx context.Context, id string, input model.SnippetInput) (*model.Snippet, error) {
	existing, err := s.repo.GetByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}

	normalized, err := s.validateInput(input)
	if err != nil {
		return nil, err
	}

	existing.Title = normalized.Title
	existing.Description = normalized.Description
	existing.Language = normalized.Language
	existing.Content = normalized.Content
	existing.IsPublic = normalized.IsPublic
	existing.UpdatedAt = s.now().UTC()

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *SnippetService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, strings.TrimSpace(id))
}

func IsValidationError(err error) (*ValidationError, bool) {
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return validationErr, true
	}

	return nil, false
}

func (s *SnippetService) validateInput(input model.SnippetInput) (model.SnippetInput, error) {
	normalized := model.SnippetInput{
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Language:    strings.TrimSpace(strings.ToLower(input.Language)),
		Content:     input.Content,
		IsPublic:    input.IsPublic,
	}

	fieldErrors := make(map[string]string)
	if normalized.Title == "" {
		fieldErrors["title"] = "Title is required."
	}
	if len([]rune(normalized.Title)) > 120 {
		fieldErrors["title"] = "Title must be 120 characters or fewer."
	}
	if len([]rune(normalized.Description)) > 280 {
		fieldErrors["description"] = "Description must be 280 characters or fewer."
	}
	if strings.TrimSpace(normalized.Content) == "" {
		fieldErrors["content"] = "Code content is required."
	}
	if normalized.Language == "" {
		fieldErrors["language"] = "Language is required."
	} else if !isAllowedLanguage(normalized.Language) {
		fieldErrors["language"] = "Language must be one of the supported options."
	}

	if len(fieldErrors) > 0 {
		return model.SnippetInput{}, &ValidationError{Fields: fieldErrors}
	}

	return normalized, nil
}

func isAllowedLanguage(language string) bool {
	for _, candidate := range model.AllowedLanguages {
		if language == candidate {
			return true
		}
	}

	return false
}
