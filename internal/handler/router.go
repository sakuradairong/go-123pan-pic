package handler

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"imagehost/internal/config"
	"imagehost/static"
)

// AuthMiddleware 核心 API 保护护卫
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenCfg := config.GlobalConfig.APIToken
		// 如果配置文件放空了代表使用者放弃防御
		if tokenCfg == "" {
			c.Next()
			return
		}

		// 适配支持 Header 注入 (WEB端拦截器) 或者 查询字符串 (简化图床软件推送提取)
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

		// 【防护高级演练】拦截网络侧信道 Timing 定时猜测攻击 
		// (通过将比较强行推至相同的微小汇编运行时间从而防止密钥逐字枚举)
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

// ConfigureRoutes 注册所有的 API 路由和静态资源
func ConfigureRoutes(r *gin.Engine, imgHandler *ImageHandler, upHandler *UploadHandler) {
	// API 分组进行统一下挂网门，使用中间件强权控制
	api := r.Group("/api")
	api.Use(AuthMiddleware()) // 全部保护！
	{
		api.POST("/upload", upHandler.HandleUpload) // 接收表单 file 字段传图
		api.GET("/images", imgHandler.HandleList)      // 瀑布流展示列表获取
		api.DELETE("/images/:id", imgHandler.HandleDelete) // 删除图片
	}

	// 静态资源：图床首页前端，这部分无需密码即可被公网下载渲染	// 【全静态内嵌化架构升级】无须再依赖同目录下的 static 文件夹！
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
