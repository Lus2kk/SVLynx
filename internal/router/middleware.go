package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: 	[]string{"https://95.182.97.36"},
		AllowMethods: 	[]string{"POST", "GET", "OPTIONS"},
		AllowHeaders: 	[]string{"Content-Type", "Autorization"},
		ExposeHeaders:  []string{"Content-Lenght"},
		AllowCredentials: true,
	})
}
