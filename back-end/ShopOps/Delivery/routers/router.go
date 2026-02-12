package routers

import (
	controllers "ShopOps/Delivery/controllers"
	Infrastructure "ShopOps/Infrastructure"
	Repositories "ShopOps/Repositories"
	Usecases "ShopOps/Usecases"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRouter(db *mongo.Database) *gin.Engine {
	router := gin.Default()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Initialize services
	jwtService := Infrastructure.NewJWTService()
	authMiddleware := Infrastructure.AuthMiddleware(jwtService)

	// Initialize rate limit service
	rateLimitService := Infrastructure.NewRateLimitService()
	
	// Apply general rate limiting to all requests
	router.Use(rateLimitService.LimitGeneral())

	// Initialize repositories
	userRepo := Repositories.NewUserRepository(db)
	businessRepo := Repositories.NewBusinessRepository(db)
	salesRepo := Repositories.NewSalesRepository(db)
	expenseRepo := Repositories.NewExpenseRepository(db)
	inventoryRepo := Repositories.NewInventoryRepository(db)
	reportRepo := Repositories.NewReportRepository(db)
	syncRepo := Repositories.NewSyncRepository(db)

	// Initialize sync service
	syncService := Infrastructure.NewSyncService(db, salesRepo, expenseRepo, inventoryRepo, syncRepo)

	// Initialize use cases
	userUC := Usecases.NewUserUseCase(userRepo, jwtService)
	businessUC := Usecases.NewBusinessUseCase(businessRepo, userRepo)
	salesUC := Usecases.NewSalesUseCase(salesRepo, businessRepo, inventoryRepo)
	expenseUC := Usecases.NewExpenseUseCase(expenseRepo, businessRepo)
	inventoryUC := Usecases.NewInventoryUseCase(inventoryRepo, businessRepo)
	reportUC := Usecases.NewReportUseCase(reportRepo, businessRepo, Infrastructure.NewExportService())
	syncUC := Usecases.NewSyncUseCase(syncService, businessRepo, salesRepo, expenseRepo, inventoryRepo, syncRepo)

	// Initialize controllers
	userController := controllers.NewUserController(userUC)
	businessController := controllers.NewBusinessController(businessUC)
	salesController := controllers.NewSalesController(salesUC)
	expenseController := controllers.NewExpenseController(expenseUC)
	inventoryController := controllers.NewInventoryController(inventoryUC)
	reportController := controllers.NewReportController(reportUC)
	syncController := controllers.NewSyncController(syncUC)

	// Public routes
	router.POST("/api/v1/auth/register", userController.Register)
	router.POST("/api/v1/auth/login", userController.Login)
	router.POST("/api/v1/auth/refresh", userController.RefreshToken)

	// Protected routes (require authentication)
	protected := router.Group("/api/v1")
	protected.Use(authMiddleware)
	{
		// User routes
		protected.GET("/users/me", userController.GetCurrentUser)
		protected.PATCH("/users/me", userController.UpdateUser)

		// Business routes
		businessRoutes := protected.Group("/businesses")
		{
			businessRoutes.POST("", businessController.CreateBusiness)
			businessRoutes.GET("", businessController.GetBusinesses)
			businessRoutes.GET("/:businessId", businessController.GetBusiness)
			businessRoutes.PATCH("/:businessId", businessController.UpdateBusiness)
		}

		// Business-specific routes (require business ID in path)
		businessSpecific := protected.Group("/businesses/:businessId")
		businessSpecific.Use(Infrastructure.BusinessMiddleware())
		{
			// Sales routes
			salesRoutes := businessSpecific.Group("/sales")
			{
				salesRoutes.POST("", salesController.CreateSale)
				salesRoutes.GET("", salesController.GetSales)
				salesRoutes.GET("/summary", salesController.GetSalesSummary)
				salesRoutes.GET("/stats", salesController.GetSalesStats)
				salesRoutes.GET("/:saleId", salesController.GetSale)
				salesRoutes.PATCH("/:saleId", salesController.UpdateSale)
				salesRoutes.DELETE("/:saleId", salesController.VoidSale)
			}

			// Expense routes
			expenseRoutes := businessSpecific.Group("/expenses")
			{
				expenseRoutes.POST("", expenseController.CreateExpense)
				expenseRoutes.GET("", expenseController.GetExpenses)
				expenseRoutes.GET("/summary", expenseController.GetExpenseSummary)
				expenseRoutes.GET("/categories", expenseController.GetExpenseCategories)
				expenseRoutes.GET("/:expenseId", expenseController.GetExpense)
				expenseRoutes.PATCH("/:expenseId", expenseController.UpdateExpense)
				expenseRoutes.DELETE("/:expenseId", expenseController.VoidExpense)
			}

			// Inventory routes
			inventoryRoutes := businessSpecific.Group("/inventory")
			{
				productsRoutes := inventoryRoutes.Group("/products")
				{
					productsRoutes.POST("", inventoryController.CreateProduct)
					productsRoutes.GET("", inventoryController.GetProducts)
					productsRoutes.GET("/low-stock", inventoryController.GetLowStock)
					productsRoutes.GET("/:productId", inventoryController.GetProduct)
					productsRoutes.PATCH("/:productId", inventoryController.UpdateProduct)
					productsRoutes.DELETE("/:productId", inventoryController.DeleteProduct)
					productsRoutes.POST("/:productId/adjust", inventoryController.AdjustStock)
					productsRoutes.GET("/:productId/history", inventoryController.GetStockHistory)
				}
			}

			// Report routes
			reportRoutes := businessSpecific.Group("/reports")
			{
				reportRoutes.GET("/dashboard", reportController.GetDashboard)
				reportRoutes.GET("/sales", reportController.GetSalesReport)
				reportRoutes.GET("/expenses", reportController.GetExpensesReport)
				reportRoutes.GET("/profit", reportController.GetProfitReport)
				reportRoutes.GET("/inventory", reportController.GetInventoryReport)
				
				// Export endpoint - 10 requests per hour rate limit (ADDED)
				reportRoutes.GET("/export", 
					rateLimitService.LimitExports(), 
					reportController.ExportReport)
				
				reportRoutes.GET("/profit/summary", reportController.GetProfitSummary)
				reportRoutes.GET("/profit/trends", reportController.GetProfitTrends)
			}

			// Sync routes
			syncRoutes := businessSpecific.Group("/sync")
			{
				// Batch endpoint - 1 restore per hour per device (ADDED)
				syncRoutes.POST("/batch", 
					rateLimitService.LimitRestore(), 
					syncController.ProcessBatch)
				
				// Status endpoint - 60 requests per minute (ADDED)
				syncRoutes.GET("/status", 
					rateLimitService.LimitSync(), 
					syncController.GetSyncStatus)
			}
		}
	}

	return router
}