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

func (h *UploadHandler) HandleUpload(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无法收到图片文件: " + err.Error()})
		return
	}

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

	if !strings.HasPrefix(fileHeader.Header.Get("Content-Type"), "image/") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"code": 415, "message": "不支持的 Content-Type 类型。"})
		return
	}

	log.Printf("[上传] 接收文件: %s", fileHeader.Filename)

	fileInfo, err := h.uploadSvc.UploadFile(fileHeader)
	if err != nil {
		log.Printf("[上传] 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    fileInfo,
		"url":     fileInfo.UserSelfURL,
	})
}
