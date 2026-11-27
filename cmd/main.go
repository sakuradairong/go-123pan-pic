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

	config.InitConfig(cfgFlag)
	cfg := config.GlobalConfig

	client := pan123.NewClient(cfg.ClientID, cfg.ClientSecret)

	uploadSvc := service.NewUploadService(client, cfg.ParentFileID, cfg.CustomDomain)
	imageSvc := service.NewImageService(client, cfg.ParentFileID, cfg.CustomDomain)

	uploadHandler := handler.NewUploadHandler(uploadSvc)
	imageHandler := handler.NewImageHandler(imageSvc)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	// ShareX clients may authenticate via the query string.
	r.Use(gin.LoggerWithFormatter(func(p gin.LogFormatterParams) string {
		return fmt.Sprintf("[GIN] %s | %d | %v | %s | %s %q\n",
			p.TimeStamp.Format("2006/01/02 15:04:05"), p.StatusCode, p.Latency, p.ClientIP, p.Method, p.Request.URL.EscapedPath())
	}), gin.Recovery())
	r.MaxMultipartMemory = 10 << 20
	r.Use(handler.CORSMiddleware())

	handler.ConfigureRoutes(r, imageHandler, uploadHandler)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("服务已启动，访问 http://localhost%s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
