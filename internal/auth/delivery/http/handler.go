package http

import (
	v1 "ad_integration/internal/auth/delivery/http/v1"
	"ad_integration/internal/auth/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	serviceAuth service.AuthService
}

func NewHandler(serviceAuth *service.AuthService) *Handler {
	return &Handler{
		serviceAuth: *serviceAuth,
	}
}

// Init создает движок и запускает цепочку настроек
func (h *Handler) Init() *gin.Engine {
	engine := gin.Default()

	h.initApi(engine)

	return engine
}

func (h *Handler) initApi(engine *gin.Engine) {
	v1Handler := v1.NewHandler(h.serviceAuth)

	api := engine.Group("/api")
	{
		v1Handler.Init(api)
	}
}
