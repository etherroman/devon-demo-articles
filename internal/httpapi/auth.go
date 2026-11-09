package httpapi

import (
	"devon-demo-articles/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *Handlers) signup(c *gin.Context) {
	var body struct {
		User *service.SignupInput `json:"user"`
	}
	if !decode(c, &body) {
		return
	}
	if body.User == nil {
		fail(c, service.Invalid(map[string]string{"user": "user 객체가 필요합니다."}))
		return
	}
	user, err := h.Service.Signup(*body.User)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(201, gin.H{"user": user})
}
func (h *Handlers) login(c *gin.Context) {
	var body struct {
		User *service.LoginInput `json:"user"`
	}
	if !decode(c, &body) {
		return
	}
	if body.User == nil {
		fail(c, service.Invalid(map[string]string{"user": "user 객체가 필요합니다."}))
		return
	}
	user, err := h.Service.Login(*body.User)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(200, gin.H{"user": user})
}
