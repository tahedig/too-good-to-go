package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tahedig/too-good-to-go/backend/internal/application/dto"
	serviceinterface "github.com/tahedig/too-good-to-go/backend/internal/application/service/interface"
	handlerinterface "github.com/tahedig/too-good-to-go/backend/internal/presentation/handler/interface"
)

type healthHandler struct {
	svc serviceinterface.HealthService
}

func NewHealthHandler(svc serviceinterface.HealthService) handlerinterface.HealthHandler {
	return &healthHandler{svc: svc}
}

// Health godoc
// @Summary      Service health
// @Description  200 when the service and all its dependencies are healthy, 503 otherwise (including during shutdown).
// @Tags         system
// @Produce      json
// @Success      200 {object} dto.HealthResponse
// @Failure      503 {object} dto.HealthResponse
// @Router       /health [get]
func (h *healthHandler) Health(c *gin.Context) {
	resp := h.svc.Check(c.Request.Context())

	status := http.StatusOK
	if resp.Status != dto.HealthStatusOK {
		status = http.StatusServiceUnavailable
	}

	c.Header("Cache-Control", "no-store")
	if c.Request.Method == http.MethodHead {
		c.Status(status)
		return
	}
	c.JSON(status, resp)
}
