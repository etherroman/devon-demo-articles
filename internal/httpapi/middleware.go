package httpapi

import (
	"devon-demo-articles/internal/service"
	"github.com/gin-gonic/gin"
	"strings"
)

func (h *Handlers) optionalAuth(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if header != "" {
		parts := strings.Fields(header)
		if len(parts) != 2 || parts[0] != "Bearer" {
			fail(c, service.Unauthorized())
			return
		}
		id, err := h.Service.Authenticate(parts[1])
		if err != nil {
			fail(c, err)
			return
		}
		c.Set("userID", id)
	}
	c.Next()
}

func requireAuth(c *gin.Context) {
	if c.GetUint("userID") == 0 {
		fail(c, service.Unauthorized())
		return
	}
	c.Next()
}
