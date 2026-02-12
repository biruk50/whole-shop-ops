package controllers

import (
	"net/http"
	"strconv"

	Domain "ShopOps/Domain"
	Infrastructure "ShopOps/Infrastructure"
	Usecases "ShopOps/Usecases"

	"github.com/gin-gonic/gin"
)

type InventoryController struct {
	inventoryUC Usecases.InventoryUseCase
}

func NewInventoryController(inventoryUC Usecases.InventoryUseCase) *InventoryController {
	return &InventoryController{inventoryUC: inventoryUC}
}

// CreateProduct godoc
// @Summary      Add new product
// @Description  Create a new product with stock information
// @Tags         inventory
// @Accept       json
// @Produce      json
// @Param        businessId  path  string                       true  "Business ID"
// @Param        request     body  Domain.CreateProductRequest  true  "Product details"
// @Success      201  {object}  Domain.Product
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/inventory/products [post]
// @Security     BearerAuth
func (c *InventoryController) CreateProduct(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	var req Domain.CreateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	product, err := c.inventoryUC.CreateProduct(businessID, userID.(string), req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusCreated, product)
}

// GetProducts godoc
// @Summary      List all products
// @Description  Get products with filtering and search
// @Tags         inventory
// @Produce      json
// @Param        businessId  path    string  true   "Business ID"
// @Param        category    query   string  false  "Product category"
// @Param        status      query   string  false  "Product status"
// @Param        low_stock   query   bool    false  "Filter low stock items"
// @Param        search      query   string  false  "Search in name, SKU, barcode"
// @Param        limit       query   int     false  "Limit results"
// @Param        offset      query   int     false  "Offset results"
// @Success      200  {array}   Domain.Product
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/inventory/products [get]
// @Security     BearerAuth
func (c *InventoryController) GetProducts(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	// Parse filters
	filters := Domain.ProductFilters{}

	// Category filter
	if category := ctx.Query("category"); category != "" {
		filters.Category = &category
	}

	// Status filter
	if status := ctx.Query("status"); status != "" {
		productStatus := Domain.ProductStatus(status)
		filters.Status = &productStatus
	}

	// Low stock filter
	if lowStock := ctx.Query("low_stock"); lowStock != "" {
		if lowStock == "true" {
			trueVal := true
			filters.LowStock = &trueVal
		}
	}

	// Search filter
	if search := ctx.Query("search"); search != "" {
		filters.Search = &search
	}

	// Pagination
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	products, err := c.inventoryUC.GetProducts(businessID, filters)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, products)
}

// GetProduct godoc
// @Summary      Get product details
// @Description  Get detailed information about a specific product
// @Tags         inventory
// @Produce      json
// @Param        businessId  path  string  true  "Business ID"
// @Param        productId   path  string  true  "Product ID"
// @Success      200  {object}  Domain.Product
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/inventory/products/{productId} [get]
// @Security     BearerAuth
func (c *InventoryController) GetProduct(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	productID := ctx.Param("productId")
	if productID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Product ID is required")
		return
	}

	product, err := c.inventoryUC.GetProductByID(productID, businessID)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusNotFound, err, "")
		return
	}

	ctx.JSON(http.StatusOK, product)
}

// UpdateProduct godoc
// @Summary      Update product
// @Description  Update product information and pricing
// @Tags         inventory
// @Accept       json
// @Produce      json
// @Param        businessId  path  string                       true  "Business ID"
// @Param        productId   path  string                       true  "Product ID"
// @Param        request     body  Domain.CreateProductRequest  true  "Product update details"
// @Success      200  {object}  Domain.Product
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/inventory/products/{productId} [patch]
// @Security     BearerAuth
func (c *InventoryController) UpdateProduct(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	productID := ctx.Param("productId")
	if productID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Product ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	var req Domain.CreateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	product, err := c.inventoryUC.UpdateProduct(productID, businessID, userID.(string), req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusOK, product)
}

// DeleteProduct godoc
// @Summary      Delete product
// @Description  Soft delete a product (mark as discontinued)
// @Tags         inventory
// @Produce      json
// @Param        businessId  path  string  true  "Business ID"
// @Param        productId   path  string  true  "Product ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/inventory/products/{productId} [delete]
// @Security     BearerAuth
func (c *InventoryController) DeleteProduct(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	productID := ctx.Param("productId")
	if productID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Product ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	if err := c.inventoryUC.DeleteProduct(productID, businessID, userID.(string)); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// AdjustStock godoc
// @Summary      Manually adjust stock
// @Description  Manually adjust product stock with movement type and reason
// @Tags         inventory
// @Accept       json
// @Produce      json
// @Param        businessId  path  string                     true  "Business ID"
// @Param        productId   path  string                     true  "Product ID"
// @Param        request     body  Domain.AdjustStockRequest  true  "Stock adjustment details"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/inventory/products/{productId}/adjust [post]
// @Security     BearerAuth
func (c *InventoryController) AdjustStock(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	productID := ctx.Param("productId")
	if productID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Product ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	var req Domain.AdjustStockRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	if err := c.inventoryUC.AdjustStock(productID, businessID, userID.(string), req); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Stock adjusted successfully"})
}

// GetLowStock godoc
// @Summary      Get products below threshold
// @Description  Get products with stock below minimum threshold
// @Tags         inventory
// @Produce      json
// @Param        businessId  path    string   true   "Business ID"
// @Param        threshold   query   float64  false  "Custom threshold (optional)"
// @Success      200  {array}   Domain.Product
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/inventory/products/low-stock [get]
// @Security     BearerAuth
func (c *InventoryController) GetLowStock(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	threshold := 0.0
	if thresholdStr := ctx.Query("threshold"); thresholdStr != "" {
		if t, err := strconv.ParseFloat(thresholdStr, 64); err == nil {
			threshold = t
		}
	}

	products, err := c.inventoryUC.GetLowStock(businessID, threshold)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, products)
}

// GetStockHistory godoc
// @Summary      Get stock movement history
// @Description  Get history of stock changes for a product
// @Tags         inventory
// @Produce      json
// @Param        businessId  path  string  true   "Business ID"
// @Param        productId   path  string  true   "Product ID"
// @Param        limit       query int     false  "Limit results (default 50)"
// @Success      200  {array}   Domain.StockMovement
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/inventory/products/{productId}/history [get]
// @Security     BearerAuth
func (c *InventoryController) GetStockHistory(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	productID := ctx.Param("productId")
	if productID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Product ID is required")
		return
	}

	limit := 50
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	history, err := c.inventoryUC.GetStockHistory(productID, businessID, limit)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, history)
}
