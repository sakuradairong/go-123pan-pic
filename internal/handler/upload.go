package handler

import (
	"errors"
	"io"
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
	const maxFileSize int64 = 50 << 20
	// Allow room for multipart headers and fields while bounding temporary disk usage.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize+(1<<20))
	fileHeader, err := c.FormFile("file")
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"code": 413, "message": "上传请求过大，文件最大支持 50MB。"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无法收到图片文件: " + err.Error()})
		return
	}

	if fileHeader.Size > maxFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"code":    413,
			"message": "文件过大，最大支持 50MB。",
		})
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	validExts := map[string]string{
		".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png",
		".gif": "image/gif", ".webp": "image/webp", ".bmp": "image/bmp",
	}
	if validExts[ext] == "" {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"code":    415,
			"message": "仅支持上传图片格式（jpg/png/gif/webp/bmp）。",
		})
		return
	}

	if !strings.HasPrefix(fileHeader.Header.Get("Content-Type"), "image/") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"code": 415, "message": "不支持的 Content-Type 类型。"})
		return
	}
	// MIME 魔数校验：读取前 512 字节确认真实文件类型
	fileReader, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "无法读取文件: " + err.Error()})
		return
	}

	buf := make([]byte, 512)
	n, err := io.ReadFull(fileReader, buf)
	fileReader.Close()
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "读取文件头失败: " + err.Error()})
		return
	}

	mimeType := http.DetectContentType(buf[:n])
	if !strings.HasPrefix(mimeType, "image/") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"code":    415,
			"message": "文件真实类型不是图片（" + mimeType + "）。",
		})
		return
	}

	log.Printf("[上传] 接收文件: %s (MIME: %s)", fileHeader.Filename, mimeType)

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
