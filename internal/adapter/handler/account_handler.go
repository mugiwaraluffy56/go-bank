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

type AccountHandler struct {
	accountService service.AccountService
	validator      validator.Validator
}

func NewAccountHandler(accountService service.AccountService, validator validator.Validator) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		validator:      validator,
	}
}

func (h *AccountHandler) Create(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": apperror.ErrUnauthorized})
		return
	}

	var input entity.CreateAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperror.ErrBadRequest})
		return
	}

	if errors := h.validator.Validate(&input); len(errors) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":  apperror.ErrValidation,
			"errors": errors,
		})
		return
	}

	account, err := h.accountService.Create(c.Request.Context(), userID.(uuid.UUID), &input)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, account.ToResponse())
}

func (h *AccountHandler) GetByID(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": apperror.ErrUnauthorized})
		return
	}

	accountIDStr := c.Param("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperror.ErrBadRequest})
		return
	}

	account, err := h.accountService.GetByID(c.Request.Context(), userID.(uuid.UUID), accountID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, account.ToResponse())
}

func (h *AccountHandler) List(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": apperror.ErrUnauthorized})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	accounts, total, err := h.accountService.GetByUserID(c.Request.Context(), userID.(uuid.UUID), page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	responses := make([]*entity.AccountResponse, len(accounts))
	for i, account := range accounts {
		responses[i] = account.ToResponse()
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

func (h *AccountHandler) GetTransactions(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": apperror.ErrUnauthorized})
		return
	}

	accountIDStr := c.Param("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperror.ErrBadRequest})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	transactions, total, err := h.accountService.GetTransactions(c.Request.Context(), userID.(uuid.UUID), accountID, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	responses := make([]*entity.TransactionResponse, len(transactions))
	for i, tx := range transactions {
		responses[i] = tx.ToResponse()
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
