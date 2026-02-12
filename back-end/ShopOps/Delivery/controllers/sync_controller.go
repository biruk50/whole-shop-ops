package controllers

import (
	"net/http"

	Domain "ShopOps/Domain"
	Infrastructure "ShopOps/Infrastructure"
	Usecases "ShopOps/Usecases"

	"github.com/gin-gonic/gin"
)

type SyncController struct {
	syncUC Usecases.SyncUseCase
}

func NewSyncController(syncUC Usecases.SyncUseCase) *SyncController {
	return &SyncController{syncUC: syncUC}
}

// ProcessBatch godoc
// @Summary      Sync multiple transactions
// @Description  Process batch sync of offline transactions (sales, expenses, products)
// @Tags         sync
// @Accept       json
// @Produce      json
// @Param        businessId  path  string               true  "Business ID"
// @Param        request     body  Domain.SyncBatch     true  "Batch sync data"
// @Success      200  {object}  Domain.SyncResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/sync/batch [post]
// @Security     BearerAuth
func (c *SyncController) ProcessBatch(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	_, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	var batch Domain.SyncBatch
	if err := ctx.ShouldBindJSON(&batch); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	// Set business ID from URL parameter
	batch.BusinessID = businessID

	response, err := c.syncUC.ProcessBatch(batch)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// GetSyncStatus godoc
// @Summary      Get sync status for business
// @Description  Get synchronization status including last sync time and pending items
// @Tags         sync
// @Produce      json
// @Param        businessId  path  string  true  "Business ID"
// @Success      200  {object}  Domain.SyncStatus
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/sync/status [get]
// @Security     BearerAuth
func (c *SyncController) GetSyncStatus(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	_, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	status, err := c.syncUC.GetSyncStatus(businessID)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, status)
}

// GetLastSync godoc
// @Summary      Get last sync time for device
// @Description  Get last synchronization timestamp for specific device
// @Tags         sync
// @Produce      json
// @Param        businessId  path    string  true  "Business ID"
// @Param        device_id   query   string  true  "Device ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/sync/last-sync [get]
// @Security     BearerAuth
func (c *SyncController) GetLastSync(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	deviceID := ctx.Query("device_id")
	if deviceID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Device ID is required")
		return
	}

	lastSync, err := c.syncUC.GetLastSync(businessID, deviceID)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	if lastSync == nil {
		ctx.JSON(http.StatusOK, gin.H{"last_sync": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"last_sync": lastSync})
}
