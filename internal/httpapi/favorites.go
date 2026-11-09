package httpapi

import "github.com/gin-gonic/gin"

func (h *Handlers) favorite(c *gin.Context, add bool) {
	a, err := h.Service.Favorite(c.Param("slug"), c.GetUint("userID"), add)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(200, gin.H{"article": a})
}
func (h *Handlers) addFavorite(c *gin.Context)    { h.favorite(c, true) }
func (h *Handlers) removeFavorite(c *gin.Context) { h.favorite(c, false) }
