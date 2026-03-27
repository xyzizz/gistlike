package handler

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gistlike/internal/config"
	"gistlike/internal/service"
)

func SetupRouter(cfg config.Config, snippetService *service.SnippetService) (*gin.Engine, error) {
	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.Use(gin.Logger())

	pageHandler := NewPageHandler(cfg.AppName, snippetService)
	apiHandler := NewAPIHandler(snippetService)
	healthHandler := NewHealthHandler()

	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.AbortWithStatusJSON(500, gin.H{"message": "internal server error"})
			return
		}

		pageHandler.RenderErrorPage(c, 500, "Server error", "Something unexpected happened. Please refresh the page and try again.")
	}))

	templates, err := loadTemplates()
	if err != nil {
		return nil, fmt.Errorf("load templates: %w", err)
	}
	router.SetHTMLTemplate(templates)

	router.Static("/static", "web/static")

	router.GET("/healthz", healthHandler.Healthz)
	router.HEAD("/healthz", healthHandler.Healthz)
	router.GET("/", pageHandler.Home)
	router.HEAD("/", pageHandler.Home)
	router.GET("/snippets/new", pageHandler.NewSnippet)
	router.GET("/snippets/:id", pageHandler.ShowSnippet)
	router.GET("/snippets/:id/raw", pageHandler.ShowRawSnippet)
	router.GET("/snippets/:id/edit", pageHandler.EditSnippet)

	api := router.Group("/api")
	{
		api.GET("/snippets", apiHandler.ListSnippets)
		api.GET("/snippets/:id", apiHandler.GetSnippet)
		api.POST("/snippets", apiHandler.CreateSnippet)
		api.PUT("/snippets/:id", apiHandler.UpdateSnippet)
		api.DELETE("/snippets/:id", apiHandler.DeleteSnippet)
	}

	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"message": "route not found"})
			return
		}

		pageHandler.RenderErrorPage(c, 404, "Page not found", "The page you are trying to open does not exist.")
	})

	router.NoMethod(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(405, gin.H{"message": "method not allowed"})
			return
		}

		pageHandler.RenderErrorPage(c, 405, "Method not allowed", "That action is not supported for this page.")
	})

	return router, nil
}

func loadTemplates() (*template.Template, error) {
	funcMap := template.FuncMap{
		"formatTime": func(value time.Time) string {
			if value.IsZero() {
				return "-"
			}
			return value.Local().Format("2006-01-02 15:04")
		},
		"truncate": func(value string, max int) string {
			runes := []rune(value)
			if len(runes) <= max {
				return value
			}
			return string(runes[:max]) + "..."
		},
	}

	pattern := filepath.Join("web", "templates", "*.html")
	templates, err := template.New("").Funcs(funcMap).ParseGlob(pattern)
	if err != nil {
		return nil, err
	}

	return templates, nil
}
