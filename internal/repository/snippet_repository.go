package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gistlike/internal/model"
)

const createSnippetsTableSQL = `
CREATE TABLE IF NOT EXISTS snippets (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	language TEXT NOT NULL,
	content TEXT NOT NULL,
	is_public INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);`

const createPublicSortIndexSQL = `
CREATE INDEX IF NOT EXISTS idx_snippets_public_created_at
ON snippets(is_public, created_at DESC);`

type SnippetRepository interface {
	EnsureSchema(ctx context.Context) error
	Create(ctx context.Context, snippet *model.Snippet) error
	GetByID(ctx context.Context, id string) (*model.Snippet, error)
	ListPublic(ctx context.Context, filter model.SnippetFilter) ([]model.Snippet, error)
	Update(ctx context.Context, snippet *model.Snippet) error
	Delete(ctx context.Context, id string) error
}

type SQLiteSnippetRepository struct {
	db *sql.DB
}

func NewSQLiteSnippetRepository(db *sql.DB) *SQLiteSnippetRepository {
	return &SQLiteSnippetRepository{db: db}
}

func (r *SQLiteSnippetRepository) EnsureSchema(ctx context.Context) error {
	statements := []string{
		createSnippetsTableSQL,
		createPublicSortIndexSQL,
	}

	for _, statement := range statements {
		if _, err := r.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("exec schema statement: %w", err)
		}
	}

	return nil
}

func (r *SQLiteSnippetRepository) Create(ctx context.Context, snippet *model.Snippet) error {
	const query = `
INSERT INTO snippets (
	id, title, description, language, content, is_public, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);`

	_, err := r.db.ExecContext(
		ctx,
		query,
		snippet.ID,
		snippet.Title,
		snippet.Description,
		snippet.Language,
		snippet.Content,
		boolToInt(snippet.IsPublic),
		formatTimestamp(snippet.CreatedAt),
		formatTimestamp(snippet.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("insert snippet: %w", err)
	}

	return nil
}

func (r *SQLiteSnippetRepository) GetByID(ctx context.Context, id string) (*model.Snippet, error) {
	const query = `
SELECT id, title, description, language, content, is_public, created_at, updated_at
FROM snippets
WHERE id = ?;`

	row := r.db.QueryRowContext(ctx, query, id)
	snippet, err := scanSnippet(row.Scan)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("select snippet: %w", err)
	}

	return snippet, nil
}

func (r *SQLiteSnippetRepository) ListPublic(ctx context.Context, filter model.SnippetFilter) ([]model.Snippet, error) {
	args := []any{1}
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
SELECT id, title, description, language, content, is_public, created_at, updated_at
FROM snippets
WHERE is_public = ?`)

	if filter.Query != "" {
		queryBuilder.WriteString(`
  AND (
    LOWER(title) LIKE LOWER(?)
    OR LOWER(description) LIKE LOWER(?)
  )`)
		search := "%" + filter.Query + "%"
		args = append(args, search, search)
	}

	queryBuilder.WriteString(`
ORDER BY created_at DESC`)

	if filter.Limit > 0 {
		queryBuilder.WriteString(`
LIMIT ?`)
		args = append(args, filter.Limit)
	}

	queryBuilder.WriteString(";")

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list public snippets: %w", err)
	}
	defer rows.Close()

	snippets := make([]model.Snippet, 0)
	for rows.Next() {
		snippet, err := scanSnippet(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("scan listed snippet: %w", err)
		}
		snippets = append(snippets, *snippet)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate listed snippets: %w", err)
	}

	return snippets, nil
}

func (r *SQLiteSnippetRepository) Update(ctx context.Context, snippet *model.Snippet) error {
	const query = `
UPDATE snippets
SET title = ?, description = ?, language = ?, content = ?, is_public = ?, updated_at = ?
WHERE id = ?;`

	result, err := r.db.ExecContext(
		ctx,
		query,
		snippet.Title,
		snippet.Description,
		snippet.Language,
		snippet.Content,
		boolToInt(snippet.IsPublic),
		formatTimestamp(snippet.UpdatedAt),
		snippet.ID,
	)
	if err != nil {
		return fmt.Errorf("update snippet: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count updated snippets: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *SQLiteSnippetRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM snippets WHERE id = ?;`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete snippet: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count deleted snippets: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

type scanFunc func(dest ...any) error

func scanSnippet(scan scanFunc) (*model.Snippet, error) {
	var snippet model.Snippet
	var isPublic int
	var createdAt string
	var updatedAt string

	if err := scan(
		&snippet.ID,
		&snippet.Title,
		&snippet.Description,
		&snippet.Language,
		&snippet.Content,
		&isPublic,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, err
	}

	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	parsedUpdatedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}

	snippet.IsPublic = isPublic == 1
	snippet.CreatedAt = parsedCreatedAt
	snippet.UpdatedAt = parsedUpdatedAt

	return &snippet, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}

	return 0
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}
