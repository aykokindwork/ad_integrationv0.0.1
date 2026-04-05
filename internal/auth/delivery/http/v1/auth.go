package v1

import (
	"ad_integration/internal/auth/schema"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initAuthRoutes(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	{
		auth.POST("/login", h.login)
	}
}

func (h *Handler) login(c *gin.Context) {
	var req schema.SignIn

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty user credentials"})
		return
	}

	userLdap, err := h.services.Authenticate(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_ldap": userLdap})
}
