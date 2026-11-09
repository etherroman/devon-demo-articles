package httpapi

import (
	"devon-demo-articles/internal/repository"
	"devon-demo-articles/internal/service"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func (h *Handlers) article(c *gin.Context) {
	a, err := h.Service.Article(c.Param("slug"), c.GetUint("userID"))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(200, gin.H{"article": a})
}
func (h *Handlers) createArticle(c *gin.Context) {
	var body struct {
		Article *service.ArticleInput `json:"article"`
	}
	if !decode(c, &body) {
		return
	}
	if body.Article == nil {
		fail(c, service.Invalid(map[string]string{"article": "article 객체가 필요합니다."}))
		return
	}
	a, err := h.Service.CreateArticle(c.GetUint("userID"), *body.Article)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(201, gin.H{"article": a})
}
func (h *Handlers) updateArticle(c *gin.Context) {
	var body struct {
		Article *service.ArticleUpdate `json:"article"`
	}
	if !decode(c, &body) {
		return
	}
	if body.Article == nil {
		fail(c, service.Invalid(map[string]string{"article": "article 객체가 필요합니다."}))
		return
	}
	a, err := h.Service.UpdateArticle(c.Param("slug"), c.GetUint("userID"), *body.Article)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(200, gin.H{"article": a})
}
func (h *Handlers) deleteArticle(c *gin.Context) {
	if err := h.Service.DeleteArticle(c.Param("slug"), c.GetUint("userID")); err != nil {
		fail(c, err)
		return
	}
	c.Status(204)
}
func (h *Handlers) listArticles(c *gin.Context) {
	query := c.Request.URL.Query()
	for _, key := range []string{"limit", "offset", "tag", "author"} {
		if len(query[key]) > 1 {
			badRequest(c, "쿼리 매개변수는 한 번만 지정하세요.")
			return
		}
	}
	limit, offset := 20, 0
	for _, p := range []struct {
		name     string
		target   *int
		min, max int
	}{{"limit", &limit, 1, 100}, {"offset", &offset, 0, 1000000}} {
		if values, ok := query[p.name]; ok {
			n, err := strconv.Atoi(values[0])
			if err != nil || n < p.min || n > p.max {
				badRequest(c, "limit은 1~100, offset은 0~1,000,000의 정수여야 합니다.")
				return
			}
			*p.target = n
		}
	}
	filter := repository.ArticleFilter{Tag: strings.ToLower(strings.TrimSpace(query.Get("tag"))), Author: strings.TrimSpace(query.Get("author")), Limit: limit, Offset: offset}
	articles, count, err := h.Service.ListArticles(filter, c.GetUint("userID"))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(200, gin.H{"articles": articles, "articlesCount": count})
}
func (h *Handlers) tags(c *gin.Context) {
	tags, err := h.Service.Tags()
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(200, gin.H{"tags": tags})
}
