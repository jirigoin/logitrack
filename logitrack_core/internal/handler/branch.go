package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logitrack/core/internal/repository"
)

type BranchHandler struct {
	repo repository.BranchRepository
}

func NewBranchHandler(repo repository.BranchRepository) *BranchHandler {
	return &BranchHandler{repo: repo}
}

func (h *BranchHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/branches", h.List)
}

func (h *BranchHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, h.repo.List())
}
