package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/gobank/internal/adapter/middleware"
	"github.com/yourusername/gobank/internal/domain/entity"
	"github.com/yourusername/gobank/internal/domain/service"
	"github.com/yourusername/gobank/internal/pkg/apperror"
	"github.com/yourusername/gobank/internal/pkg/validator"
)

type UserHandler struct {
	userService service.UserService
	validator   validator.Validator
}

func NewUserHandler(userService service.UserService, validator validator.Validator) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var input entity.CreateUserInput
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

	user, err := h.userService.Register(c.Request.Context(), &input)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"full_name":  user.FullName,
			"created_at": user.CreatedAt,
		},
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var input entity.LoginInput
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

	tokens, err := h.userService.Login(c.Request.Context(), &input)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperror.ErrBadRequest})
		return
	}

	if input.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperror.ErrBadRequest})
		return
	}

	tokens, err := h.userService.RefreshToken(c.Request.Context(), input.RefreshToken)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *UserHandler) Logout(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperror.ErrBadRequest})
		return
	}

	if err := h.userService.Logout(c.Request.Context(), input.RefreshToken); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": apperror.ErrUnauthorized})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"full_name":  user.FullName,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": apperror.ErrUnauthorized})
		return
	}

	var input entity.UpdateUserInput
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

	user, err := h.userService.Update(c.Request.Context(), userID.(uuid.UUID), &input)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"full_name":  user.FullName,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}

func handleError(c *gin.Context, err error) {
	appErr := apperror.GetAppError(err)
	if appErr != nil {
		c.JSON(appErr.StatusCode, gin.H{"error": appErr})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": apperror.ErrInternalServer})
}
