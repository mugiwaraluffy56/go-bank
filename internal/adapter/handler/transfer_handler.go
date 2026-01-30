package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/gobank/internal/adapter/middleware"
	"github.com/yourusername/gobank/internal/domain/entity"
	"github.com/yourusername/gobank/internal/domain/service"
	"github.com/yourusername/gobank/internal/pkg/apperror"
	"github.com/yourusername/gobank/internal/pkg/validator"
)

type TransferHandler struct {
	transferService service.TransferService
	validator       validator.Validator
}

func NewTransferHandler(transferService service.TransferService, validator validator.Validator) *TransferHandler {
	return &TransferHandler{
		transferService: transferService,
		validator:       validator,
	}
}

func (h *TransferHandler) Create(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": apperror.ErrUnauthorized})
		return
	}

	var input entity.CreateTransferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperror.ErrBadRequest})
		return
	}

	idempotencyKey := c.GetHeader("X-Idempotency-Key")
	if idempotencyKey != "" {
		input.IdempotencyKey = idempotencyKey
	}

	if errors := h.validator.Validate(&input); len(errors) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":  apperror.ErrValidation,
			"errors": errors,
		})
		return
	}

	transfer, err := h.transferService.Create(c.Request.Context(), userID.(uuid.UUID), &input)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, transfer.ToResponse())
}

func (h *TransferHandler) GetByID(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": apperror.ErrUnauthorized})
		return
	}

	transferIDStr := c.Param("id")
	transferID, err := uuid.Parse(transferIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperror.ErrBadRequest})
		return
	}

	transfer, err := h.transferService.GetByID(c.Request.Context(), userID.(uuid.UUID), transferID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, transfer.ToResponse())
}

func (h *TransferHandler) List(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": apperror.ErrUnauthorized})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	transfers, total, err := h.transferService.GetByUserID(c.Request.Context(), userID.(uuid.UUID), page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	responses := make([]*entity.TransferResponse, len(transfers))
	for i, t := range transfers {
		responses[i] = t.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responses,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}
