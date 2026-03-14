package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logitrack/core/internal/model"
	"github.com/logitrack/core/internal/repository"
)

type UserHandler struct {
	authRepo repository.AuthRepository
}

func NewUserHandler(authRepo repository.AuthRepository) *UserHandler {
	return &UserHandler{authRepo: authRepo}
}

func (h *UserHandler) ListDrivers(c *gin.Context) {
	drivers := h.authRepo.ListByRole(model.RoleDriver)
	c.JSON(http.StatusOK, gin.H{"drivers": drivers})
}
