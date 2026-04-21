package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"imagehost/internal/service"
)

type ImageHandler struct {
	imageSvc *service.ImageService
}

func NewImageHandler(svc *service.ImageService) *ImageHandler {
	return &ImageHandler{imageSvc: svc}
}

type imageResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	URL       string `json:"url"`
	OriginURL string `json:"origin_url"`
	CreatedAt string `json:"created_at"`
}

func (h *ImageHandler) HandleList(c *gin.Context) {
	items, err := h.imageSvc.GetImageItems()
	if err != nil {
		log.Printf("查询图床列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	images := make([]imageResponse, 0, len(items))
	for _, it := range items {
		images = append(images, imageResponse{
			ID:        it.FileID,
			Name:      it.Filename,
			Size:      it.Size,
			URL:       it.UserSelfURL,
			OriginURL: it.DownloadURL,
			CreatedAt: it.CreateAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": images})
}

func (h *ImageHandler) HandleDelete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片 ID 为空"})
		return
	}

	ids := strings.Split(id, ",")
	if err := h.imageSvc.DeleteImages(ids); err != nil {
		log.Printf("删除图片 %v 失败: %v", ids, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "deleted"})
}
