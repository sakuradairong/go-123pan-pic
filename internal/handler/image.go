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

// HandleList 获取图床目录下所有图片
func (h *ImageHandler) HandleList(c *gin.Context) {
	items, err := h.imageSvc.GetImageItems()
	if err != nil {
		log.Printf("查询图床列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	// 仅返回是文件的信息 (过滤可能存在的子文件夹)
	var images []interface{}
	for _, it := range items {
		// 123pan 中 type 通常 0是文件，1是文件夹，不过创建文件时type又传的1...这里我们只输出有 size 或文件形态的
		images = append(images, gin.H{
			"id":         it.FileID,
			"name":       it.Filename,
			"size":       it.Size,
			"url":        it.UserSelfURL,
			"origin_url": it.DownloadURL,
			"created_at": it.CreateAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    images,
	})
}

// HandleDelete 删除单个图床图片
func (h *ImageHandler) HandleDelete(c *gin.Context) {
	id := c.Param("id") // 删除支持 逗号分割多选或者单个记录
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片 ID 为空"})
		return
	}

	ids := strings.Split(id, ",")
	err := h.imageSvc.DeleteImages(ids)
	if err != nil {
		log.Printf("删除图片 %v 失败: %v", ids, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "deleted",
	})
}
