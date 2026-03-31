package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsMiddleware() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins: []string{        
    "https://svlynx.site",          
    "https://www.svlynx.site",
},
        AllowMethods:     []string{"POST", "GET", "OPTIONS"},
        AllowHeaders:     []string{"Content-Type", "Authorization", "X-Refresh-Token"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
    })
}
