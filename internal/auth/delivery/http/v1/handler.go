package v1

import (
	"ad_integration/internal/auth/service"

	"github.com/gin-gonic/gin"
)

// Это ВНУТРЕННИЙ хендлер для версии 1
type Handler struct {
	services *service.AuthService
}

func NewHandler(s service.AuthService) *Handler {
	return &Handler{
		services: &s,
	}
}

// ВНИМАНИЕ: тут должен быть (h *Handler), а не (h *http.Handler)!
func (h *Handler) Init(api *gin.RouterGroup) {
	v1 := api.Group("/v1")
	{
		h.initAuthRoutes(v1)
	}
}
