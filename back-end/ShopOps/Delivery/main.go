package main

import (
	"log"
	"os"

	routers "ShopOps/Delivery/routers"
	Infrastructure "ShopOps/Infrastructure"
	_ "ShopOps/docs"
)

// @title           ShopOps Backend API
// @version         1.0
// @description     Backend system for shop operations management.
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.email  support@shopops.com
// @license.name   Apache 2.0
// @license.url    http://www.apache.org/licenses/LICENSE-2.0.html
// @host           localhost:8080
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
func main() {
	// Initialize MongoDB
	if err := Infrastructure.InitMongo(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer Infrastructure.CloseMongo()

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Setup router using GetDB()
	router := routers.SetupRouter(Infrastructure.GetDB())

	// Start server
	log.Printf("ShopOps Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
