package handler

import (
	"biblioteca-api/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes mantiene intactas las rutas originales y añade el módulo de laptops.
func SetupRoutes(
	router *gin.Engine,
	authHandler *AuthHandler,
	bookHandler *BookHandler,
	loanHandler *LoanHandler,
	laptopHandler *LaptopHandler,
	laptopLoanHandler *LaptopLoanHandler,
	jwtSecret string,
) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	router.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })

	api := router.Group("/api")
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	protected := api.Group("")
	protected.Use(middleware.AuthRequired(jwtSecret))
	{
		protected.GET("/auth/me", authHandler.Me)
		protected.GET("/books", bookHandler.List)
		protected.GET("/books/:id", bookHandler.Get)
		protected.POST("/loans", loanHandler.Borrow)
		protected.GET("/loans/me", loanHandler.MyLoans)
		protected.PUT("/loans/:id/return", loanHandler.Return)

		protected.GET("/laptops", laptopHandler.List)
		protected.POST("/laptop-loans", laptopLoanHandler.Borrow)
		protected.GET("/laptop-loans/me", laptopLoanHandler.MyLoans)
		protected.PUT("/laptop-loans/:id/return", laptopLoanHandler.Return)
	}

	admin := api.Group("")
	admin.Use(middleware.AuthRequired(jwtSecret), middleware.AdminRequired())
	{
		admin.POST("/books", bookHandler.Create)
		admin.PUT("/books/:id", bookHandler.Update)
		admin.DELETE("/books/:id", bookHandler.Delete)
		admin.GET("/loans", loanHandler.AllLoans)
		admin.POST("/laptops", laptopHandler.Create)
	}
}
