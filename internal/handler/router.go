package handler

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"imagehost/internal/config"
	"imagehost/static"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenCfg := config.GlobalConfig.APIToken
		if tokenCfg == "" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		queryToken := c.Query("token")

		var clientToken string
		if strings.HasPrefix(authHeader, "Bearer ") {
			clientToken = strings.TrimPrefix(authHeader, "Bearer ")
		} else if authHeader != "" {
			clientToken = authHeader
		} else {
			clientToken = queryToken
		}

		// constant-time compare prevents timing-based token enumeration
		if subtle.ConstantTimeCompare([]byte(clientToken), []byte(tokenCfg)) != 1 {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "未经授权访问：请提供正确的 api_token 以进行操作。",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func ConfigureRoutes(r *gin.Engine, imgHandler *ImageHandler, upHandler *UploadHandler) {
	api := r.Group("/api")
	api.Use(AuthMiddleware())
	{
		api.POST("/upload", upHandler.HandleUpload)
		api.GET("/images", imgHandler.HandleList)
		api.DELETE("/images/:id", imgHandler.HandleDelete)
	}

	r.StaticFS("/static", http.FS(static.FS))
	r.GET("/", func(c *gin.Context) {
		htmlData, err := static.FS.ReadFile("index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "内置静态首页打包遗失")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", htmlData)
	})
}
