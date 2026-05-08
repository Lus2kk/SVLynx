package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsMiddleware() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins: []string{
    "http://localhost:5173",      
    "https://svlynx.site",          
    "https://www.svlynx.site",
    },
        AllowMethods:     []string{"POST", "GET", "OPTIONS", "PATCH", "DELETE"},
        AllowHeaders:     []string{
        "Origin",
		"Content-Length",
		"Content-Type",
		"Authorization", 
		"Accept",
		"X-Requested-With",
		"X-Refresh-Token",
    },
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
         AllowWebSockets:  true,
    })
}
