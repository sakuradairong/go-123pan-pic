package handler

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"imagehost/internal/service"
)

type UploadHandler struct {
	uploadSvc *service.UploadService
}

func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{uploadSvc: svc}
}

// HandleUpload  接受 /api/upload 的 POST 表单请求
func (h *UploadHandler) HandleUpload(c *gin.Context) {
	// 'file' 是约定好的 form 表单 file 键值，支持 ShareX / PicGo 等
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无法收到图片文件: " + err.Error(),
		})
		return
	}

	// 【安全防护1】：强行判定后缀格式，禁止脚本上传渗透
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	validExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, 
		".gif": true, ".webp": true, ".svg": true, ".bmp": true,
	}
	if !validExts[ext] {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"code":    415,
			"message": "仅支持上传图片格式（jpg/png/gif/webp/svg/bmp）。",
		})
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"code":    415,
			"message": "不支持的 Content-Type 类型。",
		})
		return
	}

	log.Printf("准备上传文件: %s", fileHeader.Filename)

	fileInfo, err := h.uploadSvc.UploadFile(fileHeader)
	if err != nil {
		log.Printf("上传业务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	// 适配外部图床响应。ShareX/PicGo 中可以通过 json_path 指定取 url
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    fileInfo,
		"url":     fileInfo.UserSelfURL, 
	})
}
