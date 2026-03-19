package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logitrack/core/internal/middleware"
	"github.com/logitrack/core/internal/model"
	"github.com/logitrack/core/internal/service"
)

type DriverHandler struct {
	routeSvc *service.RouteService
}

func NewDriverHandler(routeSvc *service.RouteService) *DriverHandler {
	return &DriverHandler{routeSvc: routeSvc}
}

func (h *DriverHandler) GetRoute(c *gin.Context) {
	user := c.MustGet(middleware.UserKey).(model.User)
	route, shipments, err := h.routeSvc.GetTodayRoute(user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no route assigned for today"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"route": route, "shipments": shipments})
}
