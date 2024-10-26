package routes

import (
	"go-jwt/controllers"
	"go-jwt/middleware"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	// Apply OpenTelemetry tracing middleware
	incomingRoutes.Use(middleware.Trace())
	
	incomingRoutes.POST("/signup", controllers.SignUp)
	incomingRoutes.POST("/signin", controllers.SignIn)
	// incomingRoutes.POST("/auth", client.Auth)
	incomingRoutes.POST("/upload", middleware.RequireAuth, controllers.UploadImages)
	incomingRoutes.GET("/validatetoken", middleware.RequireAuth, controllers.Validate)
	incomingRoutes.POST("/deleteaccount", middleware.RequireAuth, controllers.DeleteUser)
	incomingRoutes.POST("/reenable-user", controllers.ReenableUser)
}
