package httpapi

import (
	"devon-demo-articles/internal/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

func (h *Handlers) comments(c *gin.Context) {
	comments, err := h.Service.Comments(c.Param("slug"))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(200, gin.H{"comments": comments})
}
func commentBody(c *gin.Context) (*service.CommentInput, bool) {
	var body struct {
		Comment *service.CommentInput `json:"comment"`
	}
	if !decode(c, &body) {
		return nil, false
	}
	if body.Comment == nil {
		fail(c, service.Invalid(map[string]string{"comment": "comment 객체가 필요합니다."}))
		return nil, false
	}
	return body.Comment, true
}
func (h *Handlers) createComment(c *gin.Context) {
	in, ok := commentBody(c)
	if !ok {
		return
	}
	comment, err := h.Service.CreateComment(c.Param("slug"), c.GetUint("userID"), *in)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(201, gin.H{"comment": comment})
}
func commentID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		badRequest(c, "댓글 ID는 양의 정수여야 합니다.")
		return 0, false
	}
	return uint(id), true
}
func (h *Handlers) updateComment(c *gin.Context) {
	id, ok := commentID(c)
	if !ok {
		return
	}
	in, ok := commentBody(c)
	if !ok {
		return
	}
	comment, err := h.Service.UpdateComment(c.Param("slug"), id, c.GetUint("userID"), *in)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(200, gin.H{"comment": comment})
}
func (h *Handlers) deleteComment(c *gin.Context) {
	id, ok := commentID(c)
	if !ok {
		return
	}
	if err := h.Service.DeleteComment(c.Param("slug"), id, c.GetUint("userID")); err != nil {
		fail(c, err)
		return
	}
	c.Status(204)
}
