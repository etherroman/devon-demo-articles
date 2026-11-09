// Package httpapi owns routing, wire DTOs and HTTP status translation.
package httpapi

import (
	"fmt"
	"net/http"

	"devon-demo-articles/internal/service"
	"github.com/gin-gonic/gin"
)

type Handlers struct{ Service *service.Service }

func New(s *service.Service) http.Handler {
	h := &Handlers{Service: s}
	r := gin.New()
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) { fail(c, fmt.Errorf("request panic: %v", recovered)) }))
	r.HandleMethodNotAllowed = true
	r.NoRoute(func(c *gin.Context) {
		fail(c, &service.Error{Status: 404, Code: "not_found", Message: "경로를 찾을 수 없습니다."})
	})
	r.NoMethod(func(c *gin.Context) {
		fail(c, &service.Error{Status: 405, Code: "method_not_allowed", Message: "지원하지 않는 메서드입니다."})
	})
	r.GET("/healthz", func(c *gin.Context) {
		db, err := s.Repo.DB.DB()
		if err != nil {
			fail(c, err)
			return
		}
		if err := db.PingContext(c.Request.Context()); err != nil {
			fail(c, err)
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})
	api := r.Group("/api")
	api.Use(h.optionalAuth)
	api.POST("/users", h.signup)
	api.POST("/users/login", h.login)
	api.GET("/articles", h.listArticles)
	api.GET("/articles/:slug", h.article)
	api.GET("/articles/:slug/comments", h.comments)
	api.GET("/tags", h.tags)
	protected := api.Group("")
	protected.Use(requireAuth)
	protected.POST("/articles", h.createArticle)
	protected.PUT("/articles/:slug", h.updateArticle)
	protected.DELETE("/articles/:slug", h.deleteArticle)
	protected.POST("/articles/:slug/comments", h.createComment)
	protected.PUT("/articles/:slug/comments/:id", h.updateComment)
	protected.DELETE("/articles/:slug/comments/:id", h.deleteComment)
	protected.POST("/articles/:slug/favorite", h.addFavorite)
	protected.DELETE("/articles/:slug/favorite", h.removeFavorite)
	return r
}
