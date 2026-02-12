package controllers

import (
	"net/http"

	Domain "ShopOps/Domain"
	Infrastructure "ShopOps/Infrastructure"
	Usecases "ShopOps/Usecases"

	"github.com/gin-gonic/gin"
)

type BusinessController struct {
	businessUC Usecases.BusinessUseCase
}

func NewBusinessController(businessUC Usecases.BusinessUseCase) *BusinessController {
	return &BusinessController{businessUC: businessUC}
}

// CreateBusiness godoc
// @Summary      Create a new business
// @Description  Create a business profile for the authenticated user
// @Tags         businesses
// @Accept       json
// @Produce      json
// @Param        request  body  Domain.CreateBusinessRequest  true  "Business details"
// @Success      201  {object}  Domain.Business
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses [post]
// @Security     BearerAuth
func (c *BusinessController) CreateBusiness(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	var req Domain.CreateBusinessRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	business, err := c.businessUC.CreateBusiness(userID.(string), req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusCreated, business)
}

// GetBusinesses godoc
// @Summary      List user's businesses
// @Description  Get all businesses owned by the authenticated user
// @Tags         businesses
// @Produce      json
// @Success      200  {array}   Domain.Business
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses [get]
// @Security     BearerAuth
func (c *BusinessController) GetBusinesses(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	businesses, err := c.businessUC.GetUserBusinesses(userID.(string))
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, businesses)
}

// GetBusiness godoc
// @Summary      Get business details
// @Description  Get detailed information about a specific business
// @Tags         businesses
// @Produce      json
// @Param        businessId  path  string  true  "Business ID"
// @Success      200  {object}  Domain.Business
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId} [get]
// @Security     BearerAuth
func (c *BusinessController) GetBusiness(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	business, err := c.businessUC.GetBusinessByID(businessID)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusNotFound, err, "")
		return
	}

	ctx.JSON(http.StatusOK, business)
}

// UpdateBusiness godoc
// @Summary      Update business settings
// @Description  Update business profile and settings
// @Tags         businesses
// @Accept       json
// @Produce      json
// @Param        businessId  path  string  true  "Business ID"
// @Param        request     body  Domain.UpdateBusinessRequest  true  "Business update details"
// @Success      200  {object}  Domain.Business
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId} [patch]
// @Security     BearerAuth
func (c *BusinessController) UpdateBusiness(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	var req Domain.UpdateBusinessRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	business, err := c.businessUC.UpdateBusiness(businessID, userID.(string), req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusOK, business)
}
