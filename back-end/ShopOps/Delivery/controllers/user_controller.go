package controllers

import (
	"net/http"

	Domain "ShopOps/Domain"
	Infrastructure "ShopOps/Infrastructure"
	Usecases "ShopOps/Usecases"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userUC Usecases.UserUseCase
}

func NewUserController(userUC Usecases.UserUseCase) *UserController {
	return &UserController{userUC: userUC}
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account with phone/email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  Domain.RegisterRequest  true  "User registration details"
// @Success      201  {object}  Domain.User
// @Failure      400  {object}  map[string]interface{}
// @Failure      409  {object}  map[string]interface{}
// @Router       /api/v1/auth/register [post]
func (c *UserController) Register(ctx *gin.Context) {
	var req Domain.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	user, err := c.userUC.Register(req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

// Login godoc
// @Summary      Authenticate user
// @Description  Login with phone and password, receive JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  Domain.LoginRequest  true  "Login credentials"
// @Success      200  {object}  Domain.LoginResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/auth/login [post]
func (c *UserController) Login(ctx *gin.Context) {
	var req Domain.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	loginResponse, err := c.userUC.Login(req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, err, "")
		return
	}

	ctx.JSON(http.StatusOK, loginResponse)
}

// RefreshToken godoc
// @Summary      Refresh JWT token
// @Description  Get a new access token using the refresh token
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]interface{} "token"
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/auth/refresh [post]
// @Security     BearerAuth
func (c *UserController) RefreshToken(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	token, err := c.userUC.RefreshToken(userID.(string))
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token})
}

// GetCurrentUser godoc
// @Summary      Get current user profile
// @Description  Retrieve authenticated user's profile information
// @Tags         users
// @Produce      json
// @Success      200  {object}  Domain.User
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/users/me [get]
// @Security     BearerAuth
func (c *UserController) GetCurrentUser(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	user, err := c.userUC.GetCurrentUser(userID.(string))
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusNotFound, err, "")
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// UpdateUser godoc
// @Summary      Update user profile
// @Description  Update current user's profile information
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body  Domain.UpdateUserRequest  true  "User update details"
// @Success      200  {object}  Domain.User
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/users/me [patch]
// @Security     BearerAuth
func (c *UserController) UpdateUser(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	var req Domain.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	user, err := c.userUC.UpdateUser(userID.(string), req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusOK, user)
}
