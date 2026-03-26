package handler

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"gistlike/internal/model"
	"gistlike/internal/repository"
	"gistlike/internal/service"
)

type PageHandler struct {
	appName        string
	snippetService *service.SnippetService
}

func NewPageHandler(appName string, snippetService *service.SnippetService) *PageHandler {
	return &PageHandler{
		appName:        appName,
		snippetService: snippetService,
	}
}

func (h *PageHandler) Home(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))

	snippets, err := h.snippetService.ListPublic(c.Request.Context(), query)
	if err != nil {
		h.RenderErrorPage(c, http.StatusInternalServerError, "Unable to load snippets", "The snippet list could not be loaded. Please try again.")
		return
	}

	h.renderPage(c, http.StatusOK, ViewData{
		PageTitle:    "Public snippets",
		PageTemplate: "index_content",
		SearchQuery:  query,
		Snippets:     snippets,
	})
}

func (h *PageHandler) NewSnippet(c *gin.Context) {
	h.renderPage(c, http.StatusOK, ViewData{
		PageTitle:    "Create snippet",
		PageTemplate: "snippet_form_content",
		Languages:    h.snippetService.AllowedLanguages(),
		FormMode:     "create",
		Scripts:      []string{"/static/js/snippet-form.js"},
	})
}

func (h *PageHandler) ShowSnippet(c *gin.Context) {
	snippet, err := h.snippetService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.RenderErrorPage(c, http.StatusNotFound, "Snippet not found", "The snippet you are looking for does not exist, or the link is invalid.")
			return
		}

		h.RenderErrorPage(c, http.StatusInternalServerError, "Unable to load snippet", "The snippet could not be loaded right now. Please try again.")
		return
	}

	h.renderPage(c, http.StatusOK, ViewData{
		PageTitle:    snippet.Title,
		PageTemplate: "snippet_detail_content",
		Snippet:      snippet,
		Scripts:      []string{"/static/js/snippet-detail.js"},
	})
}

func (h *PageHandler) ShowRawSnippet(c *gin.Context) {
	snippet, err := h.snippetService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.Data(http.StatusNotFound, "text/plain; charset=utf-8", []byte("snippet not found\n"))
			return
		}

		c.Data(http.StatusInternalServerError, "text/plain; charset=utf-8", []byte("unable to load snippet\n"))
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", `inline; filename="`+snippetFilename(snippet)+`"`)
	c.Header("X-Content-Type-Options", "nosniff")
	c.String(http.StatusOK, snippet.Content)
}

func (h *PageHandler) EditSnippet(c *gin.Context) {
	snippet, err := h.snippetService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.RenderErrorPage(c, http.StatusNotFound, "Snippet not found", "The snippet you want to edit no longer exists.")
			return
		}

		h.RenderErrorPage(c, http.StatusInternalServerError, "Unable to load snippet", "The snippet could not be loaded for editing. Please try again.")
		return
	}

	h.renderPage(c, http.StatusOK, ViewData{
		PageTitle:    "Edit " + snippet.Title,
		PageTemplate: "snippet_form_content",
		Snippet:      snippet,
		Languages:    h.snippetService.AllowedLanguages(),
		FormMode:     "edit",
		Scripts:      []string{"/static/js/snippet-form.js"},
	})
}

func (h *PageHandler) RenderErrorPage(c *gin.Context, status int, title, message string) {
	h.renderPage(c, status, ViewData{
		PageTitle:    title,
		PageTemplate: "error_content",
		ErrorCode:    status,
		ErrorTitle:   title,
		ErrorMessage: message,
	})
}

func (h *PageHandler) renderPage(c *gin.Context, status int, data ViewData) {
	data.AppName = h.appName
	c.HTML(status, "layout", data)
}

var nonFilenameChars = regexp.MustCompile(`[^a-z0-9]+`)

func snippetFilename(snippet *model.Snippet) string {
	baseName := strings.ToLower(strings.TrimSpace(snippet.Title))
	baseName = nonFilenameChars.ReplaceAllString(baseName, "-")
	baseName = strings.Trim(baseName, "-")
	if baseName == "" {
		baseName = "snippet"
	}

	return baseName + "." + snippetExtension(snippet.Language)
}

func snippetExtension(language string) string {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "go":
		return "go"
	case "python":
		return "py"
	case "javascript":
		return "js"
	case "sql":
		return "sql"
	case "json":
		return "json"
	default:
		return "txt"
	}
}
