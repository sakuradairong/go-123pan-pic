package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"imagehost/internal/config"
	"imagehost/internal/handler"
	"imagehost/internal/pan123"
	"imagehost/internal/service"
)

func main() {
	var cfgFlag string
	flag.StringVar(&cfgFlag, "c", "conf/config.yaml", "指定配置文件路径")
	flag.Parse()

	// 1. 初始化读取配置
	config.InitConfig(cfgFlag)

	// 2. 注入 123pan 客户端基础 SDK
	client := pan123.NewClient(config.GlobalConfig.ClientID, config.GlobalConfig.ClientSecret)

	// 3. 构造上层 Service 业务操作对象
	uploadSvc := service.NewUploadService(client)
	imageSvc := service.NewImageService(client)

	// 4. 组装 Web Handler 层
	uploadHandler := handler.NewUploadHandler(uploadSvc)
	imageHandler := handler.NewImageHandler(imageSvc)

	// 5. 启动 Gin 骨架
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 【安全防护】：将单次 Multipart 表单承受的最高物理占用内存限制为 10MB
	// 若超过此量则会被拒绝或直接下刷到临时文件盘，彻底免疫大发包引起的 OOM 宕机
	r.MaxMultipartMemory = 10 << 20 

	// 提供一个基础版的跨域中间件（如果后续要在其他域名环境挂载 API 小组件的话会用到）
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	handler.ConfigureRoutes(r, imageHandler, uploadHandler)

	addr := fmt.Sprintf(":%d", config.GlobalConfig.Port)
	log.Printf("==============================")
	log.Printf("123pan 私人图床服务已启动，并挂载高危防御盾牌！")
	log.Printf("请访问 http://localhost%s 浏览！", addr)
	log.Printf("==============================")

	if err := r.Run(addr); err != nil {
		log.Fatalf("Fatal: Web 服务启动暴毙了: %v", err)
	}
}
