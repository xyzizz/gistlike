package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"gistlike/internal/model"
	"gistlike/internal/repository"
	"gistlike/internal/service"
)

type APIHandler struct {
	snippetService *service.SnippetService
}

func NewAPIHandler(snippetService *service.SnippetService) *APIHandler {
	return &APIHandler{snippetService: snippetService}
}

func (h *APIHandler) ListSnippets(c *gin.Context) {
	snippets, err := h.snippetService.ListPublic(c.Request.Context(), c.Query("q"))
	if err != nil {
		writeJSONError(c, http.StatusInternalServerError, "could not load snippets", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  snippets,
		"query": c.Query("q"),
	})
}

func (h *APIHandler) GetSnippet(c *gin.Context) {
	snippet, err := h.snippetService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": snippet})
}

func (h *APIHandler) CreateSnippet(c *gin.Context) {
	var input model.SnippetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeJSONError(c, http.StatusBadRequest, "invalid JSON body", nil)
		return
	}

	snippet, err := h.snippetService.Create(c.Request.Context(), input)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.Header("Location", "/snippets/"+snippet.ID)
	c.JSON(http.StatusCreated, gin.H{"data": snippet})
}

func (h *APIHandler) UpdateSnippet(c *gin.Context) {
	var input model.SnippetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeJSONError(c, http.StatusBadRequest, "invalid JSON body", nil)
		return
	}

	snippet, err := h.snippetService.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": snippet})
}

func (h *APIHandler) DeleteSnippet(c *gin.Context) {
	if err := h.snippetService.Delete(c.Request.Context(), c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *APIHandler) writeServiceError(c *gin.Context, err error) {
	if validationErr, ok := service.IsValidationError(err); ok {
		writeJSONError(c, http.StatusUnprocessableEntity, "validation failed", validationErr.Fields)
		return
	}

	if errors.Is(err, repository.ErrNotFound) {
		writeJSONError(c, http.StatusNotFound, "snippet not found", nil)
		return
	}

	writeJSONError(c, http.StatusInternalServerError, "internal server error", nil)
}

func writeJSONError(c *gin.Context, status int, message string, fields map[string]string) {
	c.JSON(status, gin.H{
		"message": message,
		"fields":  fields,
	})
}
